import { Injectable, UnauthorizedException } from '@nestjs/common';
import * as bcrypt from 'bcryptjs';
import { Request, Response } from 'express';
import { CookieService } from 'src/cookie/cookie.service';
import { TokenService } from 'src/token/token.service';
import { AdminService } from '@/src/admin/admin.service';

@Injectable()
export class AuthService {
  constructor(
    private adminService: AdminService,
    private tokenService: TokenService,
    private cookieService: CookieService,
  ) {}

  async validateUser(name: string, pass: string): Promise<any> {
    const user = await this.adminService.findByName(name);

    if (!user) return null;

    const isMatch = await bcrypt.compare(pass, user.password);

    if (user && isMatch) {
      const { password, ...result } = user;
      return result;
    }
    return null;
  }

  async refresh(req: Request, res: Response) {
    const refreshToken = this.cookieService.get(req, 'metric_refresh_token');
    if (!refreshToken) {
      throw new UnauthorizedException('No refresh token provided');
    }

    const accessToken = await this.tokenService.refresh(refreshToken);

    if (accessToken) {
      this.cookieService.set(res, 'metric_access_token', accessToken);
      return true;
    }

    return false;
  }

  async login(res: Response, user: any, remember: boolean) {
    const { password: _, ...payload } = user;
    const { accessToken, refreshToken } = await this.tokenService.generate(
      payload,
      remember,
    );

    this.cookieService.set(res, 'metric_access_token', accessToken);

    if (remember && refreshToken)
      this.cookieService.set(res, 'metric_refresh_token', refreshToken);
  }

  async logout(res: Response) {
    this.cookieService.set(res, 'metric_access_token', '');
    this.cookieService.set(res, 'metric_refresh_token', '');
  }
}

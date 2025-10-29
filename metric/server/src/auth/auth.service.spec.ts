import { DeepMocked, createMock } from '@golevelup/ts-jest';
import * as bcrypt from 'bcryptjs';
import { CookieService } from '../cookie/cookie.service';
import { TokenService } from '../token/token.service';
import { AuthService } from './auth.service';
import { AdminService } from '../admin/admin.service';

describe('AuthService', () => {
  let service: AuthService;
  let adminService: DeepMocked<AdminService>;
  let tokenService: DeepMocked<TokenService>;
  let cookieService: DeepMocked<CookieService>;

  beforeEach(() => {
    adminService = createMock<AdminService>();
    tokenService = createMock<TokenService>();
    cookieService = createMock<CookieService>();
    service = new AuthService(adminService, tokenService, cookieService);
  });

  describe('validateUser', () => {
    it('should try to find the user in the database', async () => {
      await service.validateUser('test', 'test');

      expect(adminService.findByName).toHaveBeenCalled();
    });

    it('should return null if the user is not found', async () => {
      jest.spyOn(adminService, 'findByName').mockResolvedValue(null);

      const result = await service.validateUser('test', 'test');
      expect(result).toBeNull();
    });

    it('should return null if the password is incorrect', async () => {
      jest.spyOn(bcrypt, 'compare').mockResolvedValue(false as never);
      const result = await service.validateUser('test', 'test');
      expect(result).toBeNull();
    });

    it('should return the user if the password is correct', async () => {
      jest.spyOn(bcrypt, 'compare').mockResolvedValue(true as never);
      const result = await service.validateUser('test', 'test');
      expect(result).not.toBeNull();
    });

    it("shouldn't contain the password in the result", async () => {
      jest.spyOn(bcrypt, 'compare').mockResolvedValue(true as never);
      const result = await service.validateUser('test', 'test');
      expect(result?.password).toBeUndefined();
    });
  });

  describe('refresh', () => {
    it('should get the refresh token cookie', async () => {
      await service.refresh({} as any, {} as any);

      expect(cookieService.get).toHaveBeenCalledTimes(1);
    });

    it("should throw an error if there's no refresh token cookie", async () => {
      jest.spyOn(cookieService, 'get').mockReturnValueOnce(undefined as never);
      await expect(service.refresh({} as any, {} as any)).rejects.toThrow();
    });

    it('should try to refresh the token', async () => {
      await service.refresh({} as any, {} as any);
      expect(tokenService.refresh).toHaveBeenCalledTimes(1);
    });

    it('should set the access token cookie', async () => {
      await service.refresh({} as any, {} as any);
      expect(cookieService.set).toHaveBeenCalledTimes(1);
    });

    it("should return false if the token couldn't be refreshed", async () => {
      jest
        .spyOn(tokenService, 'refresh')
        .mockResolvedValueOnce(undefined as never);
      const result = await service.refresh({} as any, {} as any);
      expect(result).toBe(false);
    });
  });

  describe('login', () => {
    it('should generate tokens', async () => {
      await service.login({} as any, {}, true);
      expect(tokenService.generate).toHaveBeenCalledTimes(1);
    });

    it('should set the access token cookie', async () => {
      await service.login({} as any, {}, false);
      expect(cookieService.set).toHaveBeenCalledTimes(1);
    });

    it('should set the refresh token cookie if remember is true', async () => {
      await service.login({} as any, {}, true);
      expect(cookieService.set).toHaveBeenCalledTimes(2);
    });
  });

  describe('logout', () => {
    it('should set the access and refresh token cookies to empty strings', async () => {
      const req = {} as any;
      await service.logout(req);

      expect(cookieService.set).toHaveBeenCalledWith(
        req,
        'metric_access_token',
        '',
      );
      expect(cookieService.set).toHaveBeenCalledWith(
        req,
        'metric_refresh_token',
        '',
      );
    });
  });
});

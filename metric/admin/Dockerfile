# -- Deps --
FROM node:18-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm install

# -- Base --
FROM node:18-alpine AS base
WORKDIR /usr/src/app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# -- Dev --
FROM base AS dev
EXPOSE 5173
CMD ["npm", "run", "dev"]

# -- Prod --
FROM base AS prod
RUN npm run build
EXPOSE 5173
RUN npm install -g pm2
CMD ["pm2-runtime", "start", "--name", "metric-admin", "npm", "--", "start"]

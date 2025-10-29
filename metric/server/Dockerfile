# -- Deps --
FROM node:18-alpine AS deps 
WORKDIR /app
COPY ./server/package*.json ./
RUN npm install

# -- Base --
FROM node:18-alpine AS base
WORKDIR /app/proto
COPY proto ./
WORKDIR /app/server
COPY --from=deps /app/node_modules ./node_modules
RUN apk add --no-cache openssl3
COPY ./server .
RUN npx prisma generate

# -- Dev --
FROM base AS dev
EXPOSE 4173
CMD ["npm", "run", "start:dev"]

# -- Prod --
FROM base AS prod 
RUN npm run build
EXPOSE 4173
RUN npm install -g pm2
CMD ["pm2-runtime", "start", "--name", "metric-server", "npm", "--", "start"]

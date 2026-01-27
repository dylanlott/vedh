# build stage
FROM node:14 as build-stage
WORKDIR /app
COPY package*.json ./
RUN yarn install
COPY . .
ENV NODE_ENV production
ENV VUE_APP_WEBSOCKET_URL wss://api.vedh.xyz/graphql
ENV VUE_APP_BASE_URL https://api.vedh.xyz/graphql
RUN yarn build

# production stage
FROM nginx:stable-alpine as production-stage
COPY --from=build-stage /app/dist /app
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

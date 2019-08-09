# Build the app
FROM node:10

ARG GRAPHQL_ENDPOINT
ENV REACT_APP_GRAPHQL_URI=${GRAPHQL_ENDPOINT}

RUN mkdir -p /app
WORKDIR /app

COPY package*.json ./
RUN npm install
COPY . .

RUN npm run build

# Copy built app to nginx environment
FROM nginx:stable

COPY --from=0 /app/dist /usr/share/nginx/html
COPY ./docker-nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
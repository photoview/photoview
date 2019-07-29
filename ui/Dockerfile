# Build the app
FROM node:10

ARG endpoint
ENV REACT_APP_GRAPHQL_URI=${endpoint}

RUN mkdir -p /app
WORKDIR /app

COPY package*.json ./
RUN npm install
COPY . .

RUN npm run build

# Copy built app to nginx environment
FROM nginx:stable

COPY --from=0 /app/dist /usr/share/nginx/html

EXPOSE 80
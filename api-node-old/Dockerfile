FROM node:10

ENV PRODUCTION=1

RUN mkdir -p /app
WORKDIR /app

COPY package*.json ./
RUN npm install
COPY . .

EXPOSE 4001

CMD ["npm", "start"]
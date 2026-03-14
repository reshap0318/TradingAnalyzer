FROM node:22.19.0-alpine3.22

WORKDIR /app

COPY fe/package* .

# RUN rm -rf node_modules && rm -f yarn.lock package-lock.json

RUN npm install

COPY fe .

RUN npm rebuild esbuild || true

EXPOSE 5173

CMD [ "npm", "run", "docker" ]
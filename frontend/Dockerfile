FROM node:20-alpine
WORKDIR /app
COPY frontend/package.json frontend/tsconfig.json /app/
COPY frontend/src /app/src
RUN npm install
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

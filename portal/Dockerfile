### Build web application
FROM node:16.13 as node
WORKDIR /workspace

# Copy the Node manifest, download dependencies & cache them before building and copying actual source code, so when
# source code changes, downloaded dependencies stay cached and are not downloaded again (unless manifest changes too.)
COPY package.json ./
COPY package-lock.json ./
RUN npm ci

# Now build the actual web application
COPY public ./public
COPY src ./src
COPY tsconfig.json ./
RUN npm run build

### Build portal server
FROM nginx:1.21
WORKDIR /usr/share/nginx/html
# TODO: COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=node /workspace/build/ ./
# TODO: install Amplify for metrics collection

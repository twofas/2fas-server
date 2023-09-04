FROM swaggerapi/swagger-ui

ENV URLS_PRIMARY_NAME "RestAPI"
ENV URLS "[{url: 'doc/rest.yaml', name: 'RestAPI'}, {url: 'doc/websocket.yaml', name: 'WebsocketAPI'}, {url: 'doc/admin.yaml', name: 'Admin RestAPI'},]"

COPY ./api/openapi /usr/share/nginx/html/doc

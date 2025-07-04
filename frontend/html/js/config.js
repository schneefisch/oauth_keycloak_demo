// Configuration file for service endpoints
angular.module('eventsApp.config', [])
    .constant('CONFIG', {
        // Base URIs for services
        BACKEND_URL: 'http://localhost:8082',
        KEYCLOAK_URL: 'http://localhost:8081',
        
        // Keycloak client configuration
        KEYCLOAK_REALM: 'events',
        KEYCLOAK_CLIENT_ID: 'events-frontend'
    });
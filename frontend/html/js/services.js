angular.module('eventsApp.services', ['eventsApp.config'])
    .factory('AuthService', function($http, $window, CONFIG) {
        let service = {};
        let keycloakUrl = CONFIG.KEYCLOAK_URL; // Use Keycloak URL from config
        let clientId = CONFIG.KEYCLOAK_CLIENT_ID; // Public client with PKCE
        let redirectUri = window.location.origin; // Current origin
        let tokenStorage = {}; // In-memory storage for tokens

        // Generate a random string for PKCE code_verifier
        function generateRandomString(length) {
            let text = '';
            let possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
            for (let i = 0; i < length; i++) {
                text += possible.charAt(Math.floor(Math.random() * possible.length));
            }
            return text;
        }

        // Create SHA-256 hash of the code_verifier for code_challenge
        async function sha256(plain) {
            const encoder = new TextEncoder();
            const data = encoder.encode(plain);
            const hash = await crypto.subtle.digest('SHA-256', data);
            return base64UrlEncode(hash);
        }

        // Base64Url encode the hash
        function base64UrlEncode(buffer) {
            return $window.btoa(String.fromCharCode.apply(null, new Uint8Array(buffer)))
                .replace(/\+/g, '-')
                .replace(/\//g, '_')
                .replace(/=+$/, '');
        }

        // Initialize login process
        service.login = async function() {
            // Generate code_verifier and code_challenge
            let codeVerifier = generateRandomString(64);
            let codeChallenge = await sha256(codeVerifier);

            // Store code_verifier for later use
            $window.sessionStorage.setItem('code_verifier', codeVerifier);

            // Build authorization URL
            // Redirect to Keycloak
            $window.location.href = keycloakUrl + '/realms/' + CONFIG.KEYCLOAK_REALM + '/protocol/openid-connect/auth' +
                '?client_id=' + clientId +
                '&redirect_uri=' + encodeURIComponent(redirectUri) +
                '&response_type=code' +
                '&scope=openid events-api-access' +
                '&code_challenge=' + codeChallenge +
                '&code_challenge_method=S256';
        };

        // Handle the authorization code callback
        service.handleCallback = function() {
            let urlParams = new URLSearchParams($window.location.search);
            let code = urlParams.get('code');

            if (code) {
                let codeVerifier = $window.sessionStorage.getItem('code_verifier');

                if (codeVerifier) {
                    // Exchange code for tokens
                    // Create form data properly for x-www-form-urlencoded
                    let formData = new URLSearchParams();
                    formData.append('grant_type', 'authorization_code');
                    formData.append('client_id', clientId);
                    formData.append('code', code);
                    formData.append('redirect_uri', redirectUri);
                    formData.append('code_verifier', codeVerifier);

                    return $http({
                        method: 'POST',
                        url: keycloakUrl + '/realms/' + CONFIG.KEYCLOAK_REALM + '/protocol/openid-connect/token',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded'
                        },
                        data: formData.toString()
                    }).then(function(response) {
                        // Store tokens
                        tokenStorage.accessToken = response.data.access_token;
                        tokenStorage.refreshToken = response.data.refresh_token;
                        tokenStorage.idToken = response.data.id_token;
                        tokenStorage.expiresAt = Date.now() + (response.data.expires_in * 1000);

                        // Clean up URL and session storage
                        $window.history.replaceState({}, document.title, $window.location.pathname);
                        $window.sessionStorage.removeItem('code_verifier');

                        return response.data;
                    });
                }
            }

            return Promise.resolve(null);
        };

        // Check if user is authenticated
        service.isAuthenticated = function() {
            return tokenStorage.accessToken && tokenStorage.expiresAt > Date.now();
        };

        // Get the access token
        service.getAccessToken = function() {
            if (service.isAuthenticated()) {
                return tokenStorage.accessToken;
            }
            return null;
        };

        // Logout
        service.logout = function() {
            let logoutUrl = keycloakUrl + '/realms/' + CONFIG.KEYCLOAK_REALM + '/protocol/openid-connect/logout' +
                '?redirect_uri=' + encodeURIComponent(redirectUri);

            // Clear tokens
            tokenStorage = {};

            // Redirect to Keycloak logout
            $window.location.href = logoutUrl;
        };

        return service;
    })
    .factory('EventsService', function($http, AuthService, CONFIG) {
        let service = {};
        let backendUrl = CONFIG.BACKEND_URL; // Use backend URL from config

        // Helper function to add authorization header
        function getAuthHeaders() {
            let headers = {};
            let token = AuthService.getAccessToken();
            if (token) {
                headers.Authorization = 'Bearer ' + token;
            }
            return { headers: headers };
        }

        // Fetch all events
        service.getAllEvents = function() {
            return $http.get(backendUrl + '/events', getAuthHeaders());
        };

        // Get a single event by ID
        service.getEventById = function(eventId) {
            return $http.get(backendUrl + '/events/' + eventId, getAuthHeaders());
        };

        // Create a new event
        service.createEvent = function(eventData) {
            return $http.post(backendUrl + '/events', eventData, getAuthHeaders());
        };

        // Update an existing event
        service.updateEvent = function(eventId, eventData) {
            return $http.put(backendUrl + '/events/' + eventId, eventData, getAuthHeaders());
        };

        // Delete an event
        service.deleteEvent = function(eventId) {
            return $http.delete(backendUrl + '/events/' + eventId, getAuthHeaders());
        };

        return service;
    })
    .factory('TemplateLoaderService', function($http, $compile) {
        let service = {};

        // Load a template and append it to a container
        service.loadTemplate = function(templateUrl, containerId, scope) {
            return $http.get(templateUrl)
                .then(function(response) {
                    let container = document.getElementById(containerId);
                    let compiledTemplate = $compile(response.data)(scope);
                    angular.element(container).append(compiledTemplate);
                    return compiledTemplate;
                })
                .catch(function(error) {
                    console.error('Error loading template:', error);
                    throw error;
                });
        };

        return service;
    });

angular.module('eventsApp.services', [])
    .factory('AuthService', function($http, $window) {
        var service = {};
        var keycloakUrl = 'http://localhost:8081'; // Keycloak URL
        var clientId = 'events-frontend'; // Public client with PKCE
        var redirectUri = window.location.origin; // Current origin
        var tokenStorage = {}; // In-memory storage for tokens

        // Generate a random string for PKCE code_verifier
        function generateRandomString(length) {
            var text = '';
            var possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
            for (var i = 0; i < length; i++) {
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
            var base64 = $window.btoa(String.fromCharCode.apply(null, new Uint8Array(buffer)))
                .replace(/\+/g, '-')
                .replace(/\//g, '_')
                .replace(/=+$/, '');
            return base64;
        }

        // Initialize login process
        service.login = async function() {
            // Generate code_verifier and code_challenge
            var codeVerifier = generateRandomString(64);
            var codeChallenge = await sha256(codeVerifier);

            // Store code_verifier for later use
            $window.sessionStorage.setItem('code_verifier', codeVerifier);

            // Build authorization URL
            var authUrl = keycloakUrl + '/realms/events/protocol/openid-connect/auth' +
                '?client_id=' + clientId +
                '&redirect_uri=' + encodeURIComponent(redirectUri) +
                '&response_type=code' +
                '&scope=openid' +
                '&code_challenge=' + codeChallenge +
                '&code_challenge_method=S256';

            // Redirect to Keycloak
            $window.location.href = authUrl;
        };

        // Handle the authorization code callback
        service.handleCallback = function() {
            var urlParams = new URLSearchParams($window.location.search);
            var code = urlParams.get('code');

            if (code) {
                var codeVerifier = $window.sessionStorage.getItem('code_verifier');

                if (codeVerifier) {
                    // Exchange code for tokens
                    return $http({
                        method: 'POST',
                        url: keycloakUrl + '/realms/events/protocol/openid-connect/token',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded'
                        },
                        transformRequest: function(obj) {
                            var str = [];
                            for (var p in obj) {
                                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
                            }
                            return str.join("&");
                        },
                        data: {
                            grant_type: 'authorization_code',
                            client_id: clientId,
                            code: code,
                            redirect_uri: redirectUri,
                            code_verifier: codeVerifier
                        }
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
            var logoutUrl = keycloakUrl + '/realms/events/protocol/openid-connect/logout' +
                '?redirect_uri=' + encodeURIComponent(redirectUri);

            // Clear tokens
            tokenStorage = {};

            // Redirect to Keycloak logout
            $window.location.href = logoutUrl;
        };

        return service;
    })
    .factory('EventsService', function($http, AuthService) {
        var service = {};

        // Helper function to add authorization header
        function getAuthHeaders() {
            var headers = {};
            var token = AuthService.getAccessToken();
            if (token) {
                headers.Authorization = 'Bearer ' + token;
            }
            return { headers: headers };
        }

        // Fetch all events
        service.getAllEvents = function() {
            return $http.get('/events', getAuthHeaders());
        };

        // Get a single event by ID
        service.getEventById = function(eventId) {
            return $http.get('/events/' + eventId, getAuthHeaders());
        };

        // Create a new event
        service.createEvent = function(eventData) {
            return $http.post('/events', eventData, getAuthHeaders());
        };

        // Update an existing event
        service.updateEvent = function(eventId, eventData) {
            return $http.put('/events/' + eventId, eventData, getAuthHeaders());
        };

        // Delete an event
        service.deleteEvent = function(eventId) {
            return $http.delete('/events/' + eventId, getAuthHeaders());
        };

        return service;
    })
    .factory('TemplateLoaderService', function($http, $compile) {
        var service = {};

        // Load a template and append it to a container
        service.loadTemplate = function(templateUrl, containerId, scope) {
            return $http.get(templateUrl)
                .then(function(response) {
                    var container = document.getElementById(containerId);
                    var compiledTemplate = $compile(response.data)(scope);
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

angular.module('eventsApp.services', [])
    .factory('EventsService', function($http) {
        var service = {};

        // Fetch all events
        service.getAllEvents = function() {
            return $http.get('/api/events');
        };

        // Get a single event by ID
        service.getEventById = function(eventId) {
            return $http.get('/api/events/' + eventId);
        };

        // Create a new event
        service.createEvent = function(eventData) {
            return $http.post('/api/events', eventData);
        };

        // Update an existing event
        service.updateEvent = function(eventId, eventData) {
            return $http.put('/api/events/' + eventId, eventData);
        };

        // Delete an event
        service.deleteEvent = function(eventId) {
            return $http.delete('/api/events/' + eventId);
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

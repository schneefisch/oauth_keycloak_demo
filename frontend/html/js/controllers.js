angular.module('eventsApp.controllers', ['eventsApp.services'])
    .controller('EventsController', function($scope, EventsService, TemplateLoaderService) {
        // Initialize variables
        $scope.events = [];
        $scope.loading = true;
        $scope.error = null;
        $scope.selectedEvent = null;
        $scope.loadingDetails = false;
        $scope.detailsError = null;

        // Fetch events from the API
        $scope.fetchEvents = function() {
            $scope.loading = true;
            $scope.error = null;

            EventsService.getAllEvents()
                .then(function(response) {
                    $scope.events = response.data;
                    $scope.loading = false;
                })
                .catch(function(error) {
                    console.error('Error fetching events:', error);
                    $scope.error = 'Failed to load events. Please try again later.';
                    $scope.loading = false;
                });
        };

        // Select an event and fetch its details
        $scope.selectEvent = function(eventId) {
            $scope.loadingDetails = true;
            $scope.detailsError = null;

            EventsService.getEventById(eventId)
                .then(function(response) {
                    $scope.selectedEvent = response.data;
                    $scope.loadingDetails = false;
                })
                .catch(function(error) {
                    console.error('Error fetching event details:', error);
                    $scope.detailsError = 'Failed to load event details. Please try again later.';
                    $scope.loadingDetails = false;
                });
        };

        // Clear the selected event
        $scope.clearSelectedEvent = function() {
            $scope.selectedEvent = null;
        };

        // Initialize the controller
        TemplateLoaderService.loadTemplate('templates/events-list.html', 'events-list-container', $scope)
            .then(function() {
                return TemplateLoaderService.loadTemplate('templates/event-details.html', 'event-details-container', $scope);
            })
            .then(function() {
                $scope.fetchEvents();
            })
            .catch(function(error) {
                console.error('Error initializing controller:', error);
            });
    });

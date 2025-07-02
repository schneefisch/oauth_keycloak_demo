angular.module('eventsApp.controllers', ['eventsApp.services'])
    .controller('EventsController', function($scope, EventsService, TemplateLoaderService) {
        // Initialize variables
        $scope.events = [];
        $scope.loading = true;
        $scope.error = null;

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

        // Initialize the controller
        TemplateLoaderService.loadTemplate('templates/events-list.html', 'events-list-container', $scope)
            .then(function() {
                $scope.fetchEvents();
            })
            .catch(function(error) {
                console.error('Error initializing controller:', error);
            });
    });
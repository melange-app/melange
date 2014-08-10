'use strict';

var notesControllers = angular.module('notesControllers', []);

notesControllers.controller('ListCtrl', ['$scope', function($scope) {

  melange.findMessages(["airdispat.ch/notes/body", "airdispat.ch/notes/title"], undefined, melange.angularCallback($scope, function(data) {
    $scope.notes = data;
  }));

}]);

notesControllers.controller('NewCtrl', ['$scope', '$location', function($scope, $location){

  $scope.send = function() {
    melange.createMessage({
      to: [],
      name: "notes/my-note",
      date: (new Date()).toISOString(),
      public: true,
      components: [
        {name: "airdispat.ch/notes/title", string: $scope.title},
        {name: "airdispat.ch/notes/body", string: $scope.body},
      ],
    }, melange.angularCallback($scope, function(status) {
      $location.path("/");
    }));
  };

}]);

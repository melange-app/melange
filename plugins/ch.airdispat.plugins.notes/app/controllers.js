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
      to: "",
      data: {
        "airdispat.ch/notes/title": $scope.title,
        "airdispat.ch/notes/body": $scope.body,
      }
    }, melange.angularCallback($scope, function(status) {
      $location.path("/");
    }));
  };

}]);

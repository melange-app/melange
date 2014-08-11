'use strict';

var normalize = function(str) {
  return str.toLowerCase().split(" ").join("-")
}

var notesControllers = angular.module('notesControllers', []);

notesControllers.controller('ListCtrl', ['$scope', function($scope) {

  melange.findMessages(["airdispat.ch/notes/body", "airdispat.ch/notes/title"], undefined, melange.angularCallback($scope, function(data) {
    $scope.notes = data;
  }));

}]);

notesControllers.controller('NewCtrl', ['$scope', '$location', function($scope, $location){
  $scope.send = function() {
    melange.createMessage({
      to: $scope.to,
      name: "notes/" + normalize($scope.title),
      date: (new Date()).toISOString(),
      public: $scope.to.public,
      components: {
        "airdispat.ch/notes/title": {string: $scope.title},
        "airdispat.ch/notes/body": {string: $scope.body},
      },
    }, melange.angularCallback($scope, function(status) {
      $location.path("/");
    }));
  };

}]);

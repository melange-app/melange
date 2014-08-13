'use strict';

(function() {/* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('ContactsCtrl', ['$scope', 'mlgApi',
    function($scope, mlgApi) {
      $scope.lists = mlgApi.lists()
      $scope.contacts = mlgApi.contacts()
    }]);
})();

'use strict';

function getRandomInt(min, max) {
  return Math.floor(Math.random() * (max - min)) + min;
}

(function() {/* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('ContactsCtrl', ['$scope', 'mlgApi',
    function($scope, mlgApi) {
      $scope.lists = mlgApi.lists();

      mlgApi.contacts().then(function(data) {
        $scope.contacts = data;
        if($scope.contacts.length > 0) {
          $scope.selectedContact = $scope.contacts[0];
        }
      });

      $scope.edit = false;

      var editedContacts = [];
      $scope.selectContact = function(contact) {
        if($scope.edit) {
          if(editedContacts.indexOf(contact.id) === -1) {
            editedContacts.push(contact.id);
          }
        }
        $scope.selectedContact = contact;
      }

      var save = function() {
        for(var i in editedContacts) {
          for(var k in $scope.contacts) {
            if($scope.contacts[k].id == editedContacts[i]) {
              mlgApi.updateContact($scope.contacts[k]);
              break;
            }
          }
        }
        editedContacts = [];
      }
      $scope.saveContacts = function() {
        if($scope.edit) {
          save()
        }
        $scope.edit = !$scope.edit;
      };

      $scope.subscribe = function(selectedContact) {
        selectedContact.subscribed = !selectedContact.subscribed;
        if(!$scope.edit) {
          editedContacts.push(selectedContact.id);
          save()
        }
      }

      $scope.favorite = function(selectedContact) {
        selectedContact.favorite = !selectedContact.favorite;
        if(!$scope.edit) {
          editedContacts.push(selectedContact.id);
          save()
        }
      }


      $scope.newContact = function() {
        $scope.edit = true;
        var newContact = {
          id: getRandomInt(-10000000, 0),
          addresses: [{
              alias: "",
          }],
        };
        $scope.contacts.push(newContact);
        $scope.selectContact(newContact);
      }
    }]);
})();

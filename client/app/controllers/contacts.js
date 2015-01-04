'use strict';

(function() {/* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('ContactsCtrl', ['$scope', '$timeout', 'mlgApi',
    function($scope, $timeout, mlgApi) {
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

      $scope.createList = function() {
          $scope.addListDialog = false;
      }

      $scope.createContact = function() {
          $scope.creatingContact = true;

          mlgApi.addContact({
              alias: $scope.newContactAddress,
              follow: $scope.newContactFollow,
              list: $scope.newContactList,
          }).then(function(data) {
              $scope.creatingContact = false;
              $scope.addContactDialog = false;
          });
      }

        $scope.removeContact = function(contact) {
            mlgApi.removeContact(contact).then(function(data) {
                if(data.error === false) {
                    var index = $scope.contacts.indexOf(contact);
                    $scope.contacts.splice(index, 1);

                    if($scope.contacts.length > 0) {
                        $scope.selectContact($scope.contacts[0])
                    } else {
                        $scope.selectContact({});
                    }
                    $scope.edit = false;
                }
            })
        }
    }]);
})();

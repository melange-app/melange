// Check to see if Angular is Running
(function() {
  // Common URL
  var common = "http://common.melange.127.0.0.1.xip.io:9001"
  var application = "http://app.melange.127.0.0.1.xip.io:9001"
  if (window.angular) {
    // Define UI Components
    var melangeUi = angular.module('melangeUi', []);
    // mlgToField Creates an Access Control Pane
    melangeUi.directive("mlgToField", function () {
      return {
        restrict: 'E',
        transclude: true,
        template: "<div class=\"input-group\"><span class=\"input-group-addon\"><i class=\"fa fa-group\"></i></span><input class=\"form-control\" type=\"text\" placeholder=\"Public\"><span ng-transclude></span></div>",
        link: function(elem, attr, scope) {

        }
      }
    });
  }

  function messenger(type, context) {
    window.top.postMessage({
      type: type,
      context: context,
    }, application);
  }

  var receivers = {};
  function receiveMessage(e) {
    if (e.origin !== application)
      return
    if (typeof e.data["context"] !== "object" || typeof e.data["type"] !== "string")
      return
    if (typeof receivers[e.data.type] !== "function") {
      console.log("Couldn't receive message type " + e.data.type);
      console.dir(e);
    }
    receivers[e.data.type](e.data.context)
  }
  window.addEventListener("message", receiveMessage, false);

  function callback(fn, data) {
    setTimeout(function() {
      fn(data);
    }, 0);
  }

  var melange = {
    // New Messages
    createMessage: function(msg, fn) {
      // msg is a JS Object-literal where the Keys and Values are translated to AD Directly
      receivers["createdMessage"] = function(data) {
        delete receivers["createdMessage"]
        callback(fn, data);
      }
      messenger("createMessage", msg);
    },
    // Get Messages
    findMessages: function(fields, predicate, fn) {
      receivers["foundMessages"] = function(data) {
        delete receivers["foundMessages"]
        callback(fn, data);
      }
      messenger("findMessages", {
        fields: fields,
        predicate: predicate,
      });
    },
    updateMessage: function(newMsg, id, fn) {
      // update a message with id
      receivers["updatedMessage"] = function(data) {
        delete receivers["updatedMessage"]
        callback(fn, data);
      }
      messenger("updateMessage", {
        newMsg: newMsg,
        id: id,
      });
    },
    // Remote Messages
    downloadMessage: function(addr, id, fn) {
      // will lookup a message at a specific address by name
      receivers["downloadedMessage"] = function(data) {
        delete receivers["downloadedMessage"]
        callback(fn, data);
      }
      messenger("downloadMessage", {
        addr: addr,
        id: id,
      });
    },
    downloadPublicMessages: function(fields, predicate, addr, fn) {
      // will lookup all public messages at an address
      receivers["downloadedPublicMessages"] = function(data) {
        delete receivers["downloadedPublicMessages"]
        callback(fn, data);
      }
      messenger("downloadPublicMessages", {
        addr: addr,
        fields: fields,
        predicate: predicate,
      });
    },
    // Helper Methods
    angularCallback: function(scope, fn) {
      return function(data) {
        scope.$apply(function() { fn(data) })
      };
    },
  }
  window.melange = melange;
})()


// // Example Message Object
//
// // Example Predicate
// {
//   contains: [["airdispat.ch/message/subject"]],
//   from: "",
//   to: "",
// }
//
// exports.api = melange;

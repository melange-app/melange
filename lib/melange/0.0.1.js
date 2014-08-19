// Check to see if Angular is Running
(function() {
  // Common URL
  var common = "http://common.melange:7776"
  var application = "http://app.melange:7776"

  // Angular Functions
  if (window.angular) {
    // Define UI Components
    var melangeUi = angular.module('melangeUi', []);

    melangeUi.directive('parseTo', function() {
      return {
    		restrict : 'A',
    		require: 'ngModel',
    		link: function(scope, element, attr, ngModel) {
    			function parse(string) {
            if(string === "" || string === undefined) {
              var obj = [];
              obj.public = true;
              return obj;
            } else {
              var comp = string.split(",");
              var obj = [];
              for (var i in comp) {
                obj.push({
                  alias: comp[i].alias,
                })
              }
              obj.public = false;
              return obj;
            }
    			}

    			function format(data) {
            if(data === undefined || data.length === 0) {
              return "";
            } else {
              var out = "";
              for (var i in data) {
                if(i != 0) {
                  out += ","
                }

                out += data[i].alias
              }

              return out;
            }
    			}
    			ngModel.$parsers.push(parse);
    			ngModel.$formatters.push(format);
    		}
      };
    });

    // mlgToField Creates an Access Control Pane
    melangeUi.directive("mlgToField", function () {
      return {
        require: 'ngModel',
        restrict: 'E',
        scope: { ngModel: "=" },
        template: "<div class=\"input-group\"><span class=\"input-group-addon\"><i class=\"fa fa-group\"></i></span><input ng-model=\"ngModel\" class=\"form-control\" type=\"text\" placeholder=\"Public\" parse-to></div>",
        link: function(scope, elem, attr, ngModel) {
          if(!angular.isDefined(scope.ngModel)) {
            var obj = [];
            obj.public = true;
            scope.ngModel = obj;
          } else {
            scope.ngModel.public = (scope.ngModel.length == 0)
          }
        }
      }
    });
  }

  // Wrap PostMessage
  function messenger(type, context) {
    window.top.postMessage({
      type: type,
      context: context,
    }, application);
  }

  // Receivers
  var receivers = {};
  function receiveMessage(e) {
    if (e.origin !== application)
      return
    if (typeof e.data["context"] !== "object" || typeof e.data["type"] !== "string")
      return
    if (typeof receivers[e.data.type] !== "function") {
      console.log("Couldn't receive message type " + e.data.type);
      console.dir(e);
      return
    }

    if (e.data["context"].error !== undefined && e.data["context"].error.code !== 0) {
      throw e.data["context"].error.message
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
    viewer: function(fn) {
      // msg is a JS Object-literal where the Keys and Values are translated to AD Directly
      receivers["viewerMessage"] = function(data) {
        delete receivers["viewerMessage"]
        callback(function(data) {
          fn(data);
          melange.refreshViewer();
        }, data);
      }
      setTimeout(function() {
        melange.refreshViewer(true);
      }, 0);
    },
    refreshViewer: function(sendMsg) {
      if(sendMsg === undefined) { sendMsg = false; }
      messenger("viewerUpdate",
      {
        height: document.body.scrollHeight,
        sendMsg: sendMsg,
      });
    },
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

  document.addEventListener('DOMContentLoaded', function(){
    console.log("Loaded a plugin.");
    melange.refreshViewer();
  });

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

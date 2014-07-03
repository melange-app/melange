// Wait for Inner Frame to finish Loading
document.addEventListener('DOMContentLoaded', function() {

  // API Definition
  window.melange = {
    loadScript: function(scriptName) {

    },
    log: function(logMessage) {
      messenger({
        type: "log",
        context: logMessage,
      })
    },
  }

  // The `messenger` function encapsulates methods to send messages to the mother page.
  var origin = "*";
  var messenger = function(message) {
    window.top.postMessage(message, origin);
  };

  // Try with error will execute a function inside of a try-catch block, and send a message otherwise.
  var tryWithError = function(fn) {
    try {
      fn();
    } catch (err) {
      messenger({
        type: "error",
        context: {
          message: err.message,
          destroy: true,
        },
      });
    }
  };

  // Variables for the Module
  window.module = {};

  // Different message receivers
  var receivers = {
    // Ready - load code into iFrame
    ready: function(e) {
        origin = e.origin;
        if(origin === undefined && e.context["origin"] !== undefined) { origin = e.context.origin; }

        tryWithError(function() {
          // Evaluate the Code
          window.module = e.context;
          eval(window.module["__code"]);
          delete(window.module["__code"]);

          // Activate the plugin
          if(typeof window.module.exports["activate"] === "function")
            window.module.exports.activate();

          messenger({
            type: "activated",
            context: {},
          });
        });
    },
    retrieveTemplateForAction: function(e) {
      if (e.context.action == "")
        e.context.action = window.module.exports["defaultView"];
      messenger({
        type: "template",
        context: {
          url: window.module.exports.views[e.context.action].template,
        },
      })
    },
    // Run - Start the execution of a template
    run: function(e) {
      if (typeof window.module.exports["views"] !== "object" || typeof window.module.exports["defaultView"] !== "string")
        return
      if (e.context.action == "")
        e.context.action = window.module.exports["defaultView"];
      // e.context.template
      // e.context.action
      tryWithError(function() {
        var newContext = window.module.exports.views[e.context.action].render();
        var theTemplate = Handlebars.compile(e.context.template);
        var html = theTemplate(newContext);
        messenger({
          type: "complete",
          context: {
            html: html,
          }
        });
      });
    }
  }

  // Get Ready to Receive Messages
  function receiveMessage(event) {
    // Ensure that the message is of the correct type.
    if (typeof event.data["type"] !== "string" || typeof event.data["context"] !== "object")
      return;

    // Load the correct receiver based on message type.
    if (typeof receivers[event.data.type] === "function") {
      receivers[event.data.type](event.data);
    } else {
      console.log("Couldn't determine message type.");
      console.dir(event);
    }
  }
  window.addEventListener("message", receiveMessage, false);

  // Message to signal that we are ready.
  messenger({
    type: "ready",
    context: {},
  });
});

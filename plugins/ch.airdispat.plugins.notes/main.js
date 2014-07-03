'use strict';
// ch.airdispat.plugins.notes
// Notes Plugin
// Hunter Leath

module.exports = {
  activate: function() {
    console.log("Activated notes plugin.")
  },
  deactivate: function() {},
  defaultView: "list",
  views: {
    "list": {
      template: "templates/list.html",
      render: function() {
        return {
          notes: [
            {
              author: {
                name: "Hunter Leath",
                avatar: "http://placehold.it/400x400",
                profile: "",
              },
              title: "How I met my summer",
              body: "Incididunt et dolore incididunt consequat quis ullamco veniam.",
            },
          ],
        }
      }
    }
  },
  widgets: {
    "quickNote": {
      name: "Quick Note Button",
      template: "templates/list.html",
      render: function()  {
        return {}
      }
    }
  },
}

try {
  if(window.module !== undefined) {
    $ = jQuery = module.exports;
    module.exports = {};
  }
} catch(e) { console.log("Couldn't load jQuery.") }

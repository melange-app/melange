document.addEventListener("DOMContentLoaded", function() {
  var sender = document.getElementById("sender");
  sender.onclick = function() {
    var toAddress = document.getElementById("toAddress");
    var body = document.getElementById("body");

    var parsedTo = [];

    if(toAddress.value !== "") {
      var addrs = rawTo.value.split(",");
      for (var i in addrs) {
        parsedTo.push({
          alias: addrs[i],
        });
      }
    }

    var now = new Date();

    melange.createMessage({
      to: parsedTo,
      name: "status/" + now.getTime(),
      date: now.toISOString(),
      public: true,
      components: {
        "airdispat.ch/status/body": {string: body.value},
      },
    }, function(obj) {
      console.log("Succesfully posted.");
      toAddress.value = "";
      body.value = "";
    });
  };
})

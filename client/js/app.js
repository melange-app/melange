$(document).ready(function() {
    $(".dismiss").click(function() {
      removeNotification($(this).parent());
    });


    $(".add-app").click(function () {
      $(".app-url").val($(this).attr("data-url"));
      $(".installer").removeClass("btn-primary").addClass("btn-danger").click();
    });
    // addNotification("http://cdn.macrumors.com/article-new/2013/12/app-store.jpg", "Test", "5 minutes ago", "http://google.com", 60000);
});

var show_private = 1;
var show_public = 2;

function showOnly(num) {
    if (num & show_private == show_private) {
	$(".message.mail").show();
    } else {
	$(".message.mail").hide();
    }
    if (num & show_public == show_public) {
	$(".message.subscribed").show();
    } else {
	$(".message.subscribed").hide();
    }
}

function removeNotification(elem) {
  $(elem).animate({
    right: "-370px"
    }, 750, function() {
      $(this).remove();
  });
}

function addNotification(image, text, time, url, ttl) {
  var note = $("<div/>", {class: "notification", style: "right: -370px;"})
    .append($("<a/>", { href:"#", class: "dismiss pull-right", text:"DISMISS"})
        .click(function() {
          removeNotification($(this).parent());
        }))
    .append($("<a/>", { href:url })
        .append($("<img/>", { src: image }))
        .append($("<p/>", { text: text }))
        .append($("<p/>", { class: "date", text: time })))
    .appendTo($(".nc"));
  $(note).animate({
    right: "0px"
    }, 750
  );
  window.setTimeout(function() {
    removeNotification(note);
  }, ttl*1000);
}

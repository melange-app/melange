$(document).ready(function() {
    $(".dismiss").click(function() {
      removeNotification($(this).parent());
    });

    // addNotification("http://cdn.macrumors.com/article-new/2013/12/app-store.jpg", "Test", "5 minutes ago", "http://google.com", 60000);
});

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

// <div class="notification">
//   <a href="#" class="dismiss pull-right">dismiss</a>
//   <a href="#">
//     <img src="http://cdn.macrumors.com/article-new/2013/12/app-store.jpg">
//     <p><strong>itunes.apple.com</strong> is requesting that you login</p>
//     <p class="date">5 minutes ago</p>
//   </a>
// </div>

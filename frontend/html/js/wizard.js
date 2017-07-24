$(document).ready(function() {

    $('.toggle').click(function(){
	$(this).children('i').toggleClass('fa-pencil');
        $('.form').animate({
	    height: "toggle",
	    'padding-top': 'toggle',
	    'padding-bottom': 'toggle',
	    opacity: "toggle"
        }, "slow");
    });

    $('#login').submit(function() {
	var user = $('#username').val();
	var pass = $('#password').val();
	if (!user || !pass) return;
	var data = {"Username": user, "Password": pass};
	var settings = {
	  "async": true,
	  "crossDomain": true,
	  "url": "/afw/login",
	  "method": "POST",
	  "xhrFields": { withCredentials: true },
	  "headers": {
	    "content-type": "application/json",
	    "cache-control": "no-cache"
	  },
	  "processData": false,
	  "data": JSON.stringify(data)
	};
	$.ajax(settings).done(function (response) {
	    window.location = "explorer.html";
	});

    });

    function simpleAlert(message, messageType, callback) {
        var dlg = new Alert({
            messageType: messageType || "informational",
            buttons: [{
                label: "OK",
                baseClass: "defaultButton",
                onClick: callback
            }]
        });
        dlg.setDialogContent(message);
    }
});

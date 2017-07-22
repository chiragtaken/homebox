$(document).ready(function() {

    // Toggle Function
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
	alert(user);
    });

    $.ajax({
        url: '/afw/list',
        async: false,
        success: function (data) {
		alert(data);
        },
        error: function () {
            console.log("How did we get here");
        },
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

    var cleanup = function () {
    };


});

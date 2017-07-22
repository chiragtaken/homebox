

function StartWizard(wiz, reviewer, submitter) {
      var wizId = '#' + wiz;

      $(wizId).append('<div id="wizSubmitArea"> <p style="width: 100%;text-align: center; margin-top: 10px;"> <span id="wizSubmitButton" title="Enable" class="next-button-green">Enable</span> </p> </div>');

      $(wizId).bootstrapWizard({onLast: function(tab, navigation, index) {
          debugger;
          reviewer();
        }, onTabShow: function(tab, navigation, index) {
    	var $total = navigation.find('li').length;
    	var $current = index+1;
    	var $percent = ($current/$total) * 100;
        var $pbar = navigation.find('.progress-bar')
    	$pbar.css({width:$percent+'%'});
        if ($total == $current) {
            debugger;
            $('#wizSubmitArea').show();         
            $('#wizSubmitArea').click(submitter);
        } else {
            $('#wizSubmitArea').hide();         
        }
	
    } });
    window.prettyPrint && prettyPrint()
}


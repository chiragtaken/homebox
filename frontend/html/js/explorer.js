//==================================================================
//
//      Copyright (c) 2016 by Cisco Systems, Inc.
//
//      ALL RIGHTS RESERVED. THESE SOURCE FILES ARE THE SOLE PROPERTY
//      OF CISCO SYSTEMS, Inc. AND CONTAIN CONFIDENTIAL  AND PROPRIETARY
//      INFORMATION.  REPRODUCTION OR DUPLICATION BY ANY MEANS OF ANY
//      PORTION OF THIS SOFTWARE WITHOUT PRIOR WRITTEN CONSENT OF
//      CISCO SYSTEMS, Inc. IS STRICTLY PROHIBITED.
//
//=====================================================================
stoptimeout = setTimeout(function () {

    /**
     * Test to see if the browser is Chrome
     * @type {boolean}
     */
    var IS_BROWSER_CHROME = /chrom(e|ium)/.test(navigator.userAgent.toLowerCase());


    /**
     * The root ElasticSearch url without the proxy
     * @type {string}
     */
    //var ELASTIC_SEARCH_ROOT = window.location.protocol + "//" + window.location.host + ":9200";
    var ELASTIC_SEARCH_ROOT = window.location.protocol + "//:9200";


    /**
     * The proxy URL to use
     * @type {string}
     */
    var PROXY_URL = "/cors-proxy/cn1-cors-proxy?_target=";


    /**
     * The URL of ElasticSearch. Includes port also
     * @type {string}
     */
    var ELASTIC_SEARCH_URL = PROXY_URL + ELASTIC_SEARCH_ROOT;


    /**
     * The amount of time after which to reload
     * @type {number}
     */
    var RELOAD_FILTER_TIME = 15 * 1000;


    // == Polyfills ==========================================================================

    if (!String.prototype.replaceAll) {
        String.prototype.replaceAll = function (search, replacement) {
            return this.replace(new RegExp(search, 'g'), replacement);
        };
    }

    // =======================================================================================


    var $eplDashboard = $('#epl-dashboards-iframe-container').find('iframe');
    var $kibanaIFrame = $('#kibana-iframe-container').find('iframe');
    var $kibanaIframeContent = null;
    var frameAngular  = null;
    var $eplFilters = $('.epl-filters');
    var $switchesUl = $('#epl-filter-switches');
    var $vrfsUl = $('#epl-filter-vrf');
    var $resetBtn = $('#epl-reset-button');
    var $dummyClickArea = $('#epl-dummy-click-area');


    // =======================================================================================


    /**
     * Check if the filter bar exists
     * @returns {boolean}
     */
    function doesFilterBarExists() {
        return ($kibanaIframeContent.find('filter-bar').length > 0);
    }


    /**
     * Hide the Kibana timepicker
     */
    function hideKibanaTimePicker() {
        $kibanaIframeContent.find('.navbar-timepicker .to-body').last().hide();
    }


    /**
     * Show the Kibana timepicker
     */
    function showKibanaTimePicker() {
        $kibanaIframeContent.find('.navbar-timepicker .to-body').last().show();
    }


    /**
     * Test if page is explorer screen
     */
    function isExplorerPage() {
        var hash = 'com_cisco_xmp_web_page_epl_explorer';
        return window.location.hash.indexOf(hash) > -1 ||
                (window.top && window.top.location.hash.indexOf(hash) > -1);
    }


    var makeSafeForCSSMap = {};
    // Credits: http://stackoverflow.com/a/7627603
    function makeSafeForCSS(name) {

        if (makeSafeForCSSMap[name]) {
            return makeSafeForCSSMap[name];
        }

        var value = name.replace(/[^a-z0-9]/g, function(s) {
            var c = s.charCodeAt(0);
            if (c == 32) return '-';
            if (c >= 65 && c <= 90) return '_' + s.toLowerCase();
            return '__' + ('000' + c.toString(16)).slice(-4);
        });

        makeSafeForCSSMap[name] = value;
        return value;
    }

    var filterMapElements = {};
    function getFilterDomElement(id) {
        if (filterMapElements[id]) {
            return filterMapElements[id];
        }
        filterMapElements[id] = $(id);
        return filterMapElements[id];
    }


    /**
     * Called when the filters change
     */
    function onFiltersChange(filters) {
        console.log("on filters changed", filters);

        if (filters.length == 0) {
            // set the default filters
            $eplFilters.removeClass('selected');
        } else {

            $.each(filters, function (i, filter) {

                var valueSafe = makeSafeForCSS(filter.meta.value);
                var $item;
                if (filter.meta.key == "VRF") {
                    $item = getFilterDomElement('#epl_filter_VRF_' + valueSafe);
                    if ($item.length > 0) {
                        $item.addClass('selected');
                    }
                }

                if (filter.meta.key == "Switch_Name") {
                    $item = getFilterDomElement('#epl_filter_Switch_Name_' + valueSafe);
                    if ($item.length > 0) {
                        $item.addClass('selected');
                    }
                }
            });
        }
    }


    function setAngularFrame() {
        if (!frameAngular) {
            frameAngular = $kibanaIFrame[0].contentWindow.angular;
        }
    }


    var shouldTriggerFilterChange = true;


    /**
     * Called as soon as the frame is ready
     * This porition assumes that the iframe has some code
     * to trigger events. See: http://stackoverflow.com/a/16290979
     */
    $eplDashboard.on("iframeloading", function () {

        // To access window: http://stackoverflow.com/a/1654262
        var iframe = $(this)[0];
        var iframewindow= iframe.contentWindow ? iframe.contentWindow : iframe.contentDocument.defaultView;
        var $moduleFrame = $('#epl-frame-container');
        iframewindow.EPL_CONFIG = {
            WINDOW_WIDTH: $moduleFrame.width(),
            WINDOW_HEIGHT: $moduleFrame.height(),
            TIME_ZONE: DcnmApp.about.dcnmServerTimezone
        };

        console.log("EPL_CONFIG", iframewindow.EPL_CONFIG);

        // Backwards compatibility
        if (!String.prototype.startsWith) {
            String.prototype.startsWith = function (str) {
                return !this.indexOf(str);
            }
        }

        // In Chrome, we do not need to proxy calls
        // as once the certificate is accepted, we don't
        // need to worry about port
        //if (!IS_BROWSER_CHROME) {
            // Initiate HTTP interception to use proxy for all
            // calls that will be made to ElasticSearch
            (function (XHR) {
                var open = XHR.prototype.open;
                var send = XHR.prototype.send;

                XHR.prototype.open = function (method, url, async, user, pass) {
                    if (url.startsWith(ELASTIC_SEARCH_ROOT)) {
                        url = PROXY_URL + encodeURI(url);
                    }
                    this._url = url;
                    open.call(this, method, url, async, user, pass);
                };

                XHR.prototype.send = function (data) {
                    var self = this;
                    var oldOnReadyStateChange;
                    var url = this._url;

                    function onReadyStateChange() {
                        /*if(self.readyState == 4) { // request is complete
                         }*/

                        if (oldOnReadyStateChange) {
                            oldOnReadyStateChange();
                        }
                    }

                    /* Set xhr.noIntercept to true to disable the interceptor for a particular call */
                    if (!this.noIntercept) {
                        if (this.addEventListener) {
                            this.addEventListener("readystatechange", onReadyStateChange, false);
                        } else {
                            oldOnReadyStateChange = this.onreadystatechange;
                            this.onreadystatechange = onReadyStateChange;
                        }
                    }

                    send.call(this, data);
                }
            })(iframewindow.XMLHttpRequest);
        //}
    });

    /**
     * Called when iframe is loaded
     */
    $kibanaIFrame.on('load', function() {

        // trigger click to hide menu

        // Find the reference to AngularJS
        setAngularFrame();

        // Load the contents
        $kibanaIframeContent = $kibanaIFrame.contents();

        // Initialize onFiltersChange()
        var initializeOnFiltersLengthChangeCounter = 0;
        var initializeOnFiltersLengthChange = function () {

            // Counter to quit after a while
            if (++initializeOnFiltersLengthChangeCounter >= 1000) {
                return;
            }

            // If the filter bar doesn't exist, it means
            if (!doesFilterBarExists()) {
                setTimeout(initializeOnFiltersLengthChange, 100);
                return;
            }

            // Move the filters DIV
            var $kibanaContents = $kibanaIFrame.contents();
            $kibanaContents.find('head').append( $('<link rel="stylesheet" type="text/css" />').attr('href', '/module/epl/css/epl-filters.css') );
            $eplFilters.show().appendTo($kibanaContents.find('#kibana-body .content'));

            hideKibanaTimePicker();
        };

        initializeOnFiltersLengthChange();

        // Trigger a click in the dummy area so the
        // menu can be hidden and the parent can process
        // clicks as normal
        $kibanaIframeContent.on('click', function () {
            $dummyClickArea.trigger('click');
        });

        // I know this is ugly way of doing it but it's the only
        // good and reliable solution. When the user mouses out,
        // then reset the filters
        /*$kibanaIframeContent.find('#kibana-body').mouseup(function() {

            if (shouldTriggerFilterChange == false) {
                shouldTriggerFilterChange = true;
                return;
            }

            var $filterBar = $kibanaIframeContent.find('filter-bar');
            if ($filterBar == null) { return; }

            setAngularFrame();
            var $filterBarScope = frameAngular.element($filterBar.find('.bar').first()).scope();
            if ($filterBarScope == null) { return; }

            if ($filterBarScope.filters) {
                setTimeout(function () {
                    if ($filterBarScope.filters) {
                        onFiltersChange($filterBarScope.filters);
                    }
                }, 25);
            }
        });*/
    });


    /**
     * Push a new filter
     */
    function pushFilter(item) {

        var $filterBar = $kibanaIframeContent.find('filter-bar');
        if ($filterBar == null) { return; }

        setAngularFrame();
        var $filterBarScope = frameAngular.element($filterBar.find('.bar').first()).scope();
        if ($filterBarScope == null) { return; }

        $filterBarScope.addFilters(item).then(function () {
            $filterBarScope.applyFilters();
        });
    }


    /**
     * Get a list of all the filters
     * @returns {*|Array}
     */
    function getAllFilters() {
        var filters = [];

        var $filterBar = $kibanaIframeContent.find('filter-bar');
        if ($filterBar == null) { return filters; }

        setAngularFrame();
        var $filterBarScope = frameAngular.element($filterBar.find('.bar').first()).scope();
        if ($filterBarScope == null) { return filters; }

        for (var i = 0; i < $filterBarScope.filters.length; i++) {
            var filter = $filterBarScope.filters[i];
            filters.push(JSON.parse(JSON.stringify({
                meta: filter.meta,
                query: filter.query
            })));
        }
        return filters;
    }

    /**
     * Returns the filter by its given alias
     * @param name
     * @returns {*}
     */
    function findFilterByName(name) {
        var $filterBar = $kibanaIframeContent.find('filter-bar');
        if ($filterBar == null) { return; }

        setAngularFrame();
        var $filterBarScope = frameAngular.element($filterBar.find('.bar').first()).scope();
        if ($filterBarScope == null) { return; }

        for (var i = 0; i < $filterBarScope.filters.length; i++) {
            var filter = $filterBarScope.filters[i];
            if (filter.meta.alias == name) {
                return filter;
            }
        }
        return null;
    }

    /**
     * Remove a filter
     */
    function removeFilter(item) {
        if (!item) {
            return;
        }

        var $filterBar = $kibanaIframeContent.find('filter-bar');
        if ($filterBar == null) { return; }

        setAngularFrame();
        var $filterBarScope = frameAngular.element($filterBar.find('.bar').first()).scope();
        if ($filterBarScope == null) { return; }

        $filterBarScope.removeFilter(item);
    }


    // ################################################################################################


    //$switchesUl.html('<li class="selected"><em>*</em></li>');
    //$vrfsUl.html('<li class="selected"><em>*</em></li>');

    // Build the filters drop-down lists
    var reloadVRFDrownDownList = function () {

        if (!isExplorerPage()) {
            return;
        }

        $.ajax({
            type: "POST",
            url: ELASTIC_SEARCH_URL + "/epl_cache_today/_search",
            data: JSON.stringify({
                "aggs": {
                    "tenent": {
                        "terms": {
                            "field": "VRF",
                            "order": {"_term": "asc"}
                        }
                    }
                }
            }),
            contentType: "application/json",
            dataType: "json",
            success: function (data) {
                setTimeout(reloadVRFDrownDownList, RELOAD_FILTER_TIME);
                $vrfsUl.html('');
                if (typeof data === 'string') {
                    data = JSON.parse(data);
                }
                $.each(data.aggregations.tenent.buckets, function (i, item) {
                    var id = 'epl_filter_VRF_' + makeSafeForCSS(item.key);
                    if (item.key) {
                        $vrfsUl.append('<li id="' + id + '">' + item.key + '</li>');
                    } else {
                        $vrfsUl.append('<li class="empty_selection" id="' + id + '">' + "<i style='color: red;'>NO-VRF</i>" + '</li>');
                    }
                });
            },
            error: function () {
                setTimeout(reloadVRFDrownDownList, RELOAD_FILTER_TIME);
            }
        });
    };

    reloadVRFDrownDownList();

    var reloadSwitchesDropDownList = function () {

        if (!isExplorerPage()) {
            return;
        }

        $.ajax({
            type: "POST",
            url: ELASTIC_SEARCH_URL + "/epl_cache_today/_search",
            data: JSON.stringify({
                "aggs": {
                    "tors": {
                        "terms": {
                            "field": "Switch_Name",
                            "order": {"_term": "asc"}
                        }
                    }
                }
            }),
            contentType: "application/json",
            dataType: "json",
            success: function (data) {
                setTimeout(reloadSwitchesDropDownList, RELOAD_FILTER_TIME);
                $switchesUl.html('');
                if (typeof data === 'string') {
                    data = JSON.parse(data);
                }
                $.each(data.aggregations.tors.buckets, function (i, item) {
                    var id = 'epl_filter_Switch_Name_' + makeSafeForCSS(item.key);
                    if (item.key) {
                        $switchesUl.append('<li id="' + id + '">' + item.key + '</li>');
                    } else {
                        $switchesUl.append('<li class="empty_selection" id="' + id + '">' + "<i style='color: red;'>NO-SWITCH</i>" + '</li>');
                    }
                });
            },
            error: function () {
                setTimeout(reloadSwitchesDropDownList, RELOAD_FILTER_TIME);
            }
        });
    };
    reloadSwitchesDropDownList();


    $switchesUl.on("click", "li", function() {
        shouldTriggerFilterChange = false;
        //$('li', $switchesUl).removeClass('selected');
        //$(this).addClass('selected');
        var $this = $(this);
        var text = $this.text();
        if ($this.hasClass('empty_selection')) {
            text = "";
        }

        var filter = { meta: { negate: false, index: "epl_cache_today" }, query: { match: {} } };
        filter.query.match["Switch_Name"] = { query: text, type: 'phrase' };
        pushFilter(filter);
        // see in kibana code: __webpack_require__(726)
    });

    $vrfsUl.on("click", "li", function() {
        shouldTriggerFilterChange = false;
        //$('li', $vrfsUl).removeClass('selected');
        //$(this).addClass('selected');
        var $this = $(this);
        var text = $this.text();
        if ($this.hasClass('empty_selection')) {
            text = "";
        }
        var filter = { meta: { negate: false, index: "epl_cache_today" }, query: { match: {} } };
        filter.query.match["VRF"] = { query: text, type: 'phrase' };
        pushFilter(filter);
    });

    $('#epl-filter-ipv4_ipv6').on("click", "li", function () {
        var text = $(this).text();
        var textLower = text.toLowerCase();

        removeFilter(findFilterByName("IPv4 Only"));
        removeFilter(findFilterByName("IPv6 Only"));

        if (textLower.indexOf('ipv4') > -1 && textLower.indexOf('ipv6') > -1) {
            // do nothing
        } else {
            var filter = {
                meta: {
                    negate: false,
                    index: "epl_cache_today",
                    alias: text + " Only"
                },
                "query": {
                    "bool": {
                        "must": [
                            {
                                "wildcard": {
                                    "EndpointIdentifier": text + "*"
                                }
                            }
                        ]
                    }
                }
            };
            pushFilter(filter);
        }
    });

    $eplFilters.find('.epl-input-search').on('keypress', function (e) {
        if (e.keyCode == 13 || e.which === 13) {
            var $this = $(this);
            var val = $.trim($this.val());

            // Test if it's IPv6 and if it is,
            // reduce it down to how the switches
            // save it
            if (val && val.indexOf(':') > -1) {
                // Credits: http://stackoverflow.com/questions/7043983/ipv6-address-into-compressed-form-in-java
                // Not the 0+
                val = val.replaceAll("((?::0+\\b){2,}):?(?!\\S*\\b\\1:0+\\b)(\\S*)", "::$2")
            }

            var filter = { meta: { negate: false, index: "epl_cache_today" }, query: { match: {} } };
            filter.query.match["IP"] = { query: val, type: 'phrase' };
            pushFilter(filter);
            $this.val('');
        }
    });

    $resetBtn.click(function () {
        var url = $kibanaIFrame.attr('src');
        $kibanaIFrame.contents()[0].location.hash = url.substring(url.indexOf('#') + 1);
        // Trigger a refresh
        $eplFilters.parent().find('form[name="queryInput"]').find('button[type="submit"]').click();
    });


    // ################################################################################################

    var lastTimeFilter = {
        from: "now-5y",
        to: "now",
        mode: "quick"
    };
    var default_query = null;

    var $listItems = $('.epl-navigator li');
    var oldFilters = [];
    $listItems.click(function () {

        if ($(this).hasClass('active')) {
            return;
        }

        var isNavigationFromDashboards = $('.epl-navigator li.active').index() > 1;
        var index = $listItems.index(this);

        function setHash(hash) {
            $kibanaIFrame.first()[0].contentWindow.location.hash = hash;
        }

        function getKibanaHash() {
            return $kibanaIFrame.first()[0].contentWindow.location.hash;
        }

        function showKibana() {
            $kibanaIFrame.parent().show();
            $eplDashboard.parent().hide();
        }

        function showEplDashboards() {
            $kibanaIFrame.parent().hide();
            $eplDashboard.parent().show();
            var indices = [0,1,2];
            indices.forEach(function(i) {
              $eplDashboard.eq(i).hide();
            });

            // reload Operation Heatmap iframe on tab change
            if (index == 3) {
                $eplDashboard.eq(index - 2)[0].contentWindow.location.reload();
            }

            // Show the selected dashboard
            $eplDashboard.eq(index - 2).show();
        }

        $listItems.removeClass('active');
        $(this).addClass('active');
        $eplFilters.hide();


        setAngularFrame();
        var $scope = frameAngular.element('body').scope();
        var getDiscoverSearchScope = function () {
            return frameAngular
                .element($eplFilters.parent().find('form[name="discoverSearch"]').find('button[type="submit"]')[0])
                .scope();
        };
        var getDashboardScope = function () {
            return frameAngular
                .element($eplFilters.parent().find('form[name="queryInput"]').find('button[type="submit"]')[0])
                .scope();
        };

        var getKibanaNavBarItem = function (filterItem) {
            return $eplFilters.parents('body')
                    .find('.nav.navbar-nav[role="navigation"]')
                    .find('li')
                    .filter(function() {
                        return $.trim($(this).text()) == filterItem;
                    })
                    .find('a');
        };

        var closeTimePickerIfOpen = function () {
            try {
                var $picker = $eplFilters.parents('body').find('config').filter(function () {
                    return $(this).attr('config-template') == "pickerTemplate";
                });
                // If has children means its open
                if ($picker.children().length > 0) {
                    $picker.find('.config-close.remove').click();
                }
            } catch (err) {
                console.error(err);
            }
        };

        /*if (isNavigationFromDashboards && index <= 1) {
            showKibana();
            if (index == 1) {
                $eplFilters.show();
            }
            return;
        }*/

        if (index == 0) {
            oldFilters = getAllFilters();
            if (lastTimeFilter) {
                $scope.timefilter.time = lastTimeFilter;
            }
            if (default_query && getDashboardScope()) {
                getDashboardScope().state.query = default_query;
            }

            //setHash('/discover?embed=true');
            try {
                getKibanaNavBarItem('Discover')[0].click();
            } catch (err) {}
            closeTimePickerIfOpen();
            showKibanaTimePicker();
            showKibana();
        } else if (index == 1) {

            lastTimeFilter = JSON.parse(JSON.stringify($scope.timefilter.time));
            if (getDiscoverSearchScope()) {
                default_query = getDiscoverSearchScope().state.query;
            }

            $scope.timefilter.time = {
                from: "now-5y",
                to: "now",
                mode: "quick"
            };

            var $form = $eplFilters.parent().find('form[name="discoverSearch"]');
            $form.find('input').val('');
            $form.find('button[type="submit"]').click();

            //getDiscoverSearchScope().state.query = "";
            $eplFilters.show();
            //setHash('/dashboard/epl_today_dashboard?embed=true');
            if (getKibanaHash().indexOf('epl_today_dashboard') == -1) {
                try {

                    getKibanaNavBarItem('Dashboard')[0].click();
                } catch (err) {}
                setTimeout(function () {
                    pushFilter(oldFilters);
                }, 100);
            }



            // Trigger a click to reset the search query
            /*var $form = $eplFilters.parent().find('form[name="queryInput"]');
            $form.find('input').val('');
            $form.find('button[type="submit"]').click();*/
            closeTimePickerIfOpen();
            hideKibanaTimePicker();
            showKibana();
        } else {
            showEplDashboards();
        }
    });


}, 0);


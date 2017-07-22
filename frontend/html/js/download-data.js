(function ($) {

    /**
     * The root ElasticSearch url without the proxy
     * @type {string}
     */
    //var ELASTIC_SEARCH_ROOT = window.location.protocol + "//" + window.location.host + ":9200";
    var ELASTIC_SEARCH_ROOT = window.location.protocol + "//172.28.8.158:30000";


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
     * Get the filters for download data
     */
    function getAllFilters() {

        var $kibanaIFrame = $('#kibana-iframe-container').find('iframe');
        var $filterBar = $kibanaIFrame.contents().find('filter-bar');
        if ($filterBar == null) {
            return {
                bool: {
                    'must': [],
                    'must_not': []
                }
            };
        }

        var frameAngular = $kibanaIFrame[0].contentWindow.angular;
        var $filterBarScope = frameAngular.element($filterBar.find('.bar').first()).scope();
        if ($filterBarScope == null) { return; }

        var must = [];
        var mustNot = [];

        var timeFilter = $filterBarScope.$root.$$timefilter.getBounds();

        must.push({
            "range":{
                "timestamp":{
                    "lte":timeFilter.max.valueOf(),
                    "gte":timeFilter.min.valueOf(),
                    "format":"epoch_millis"
                }
            }
        });
        must.push({
            "query":{
                "query_string":{
                    "analyze_wildcard":true,
                    "query":"*"
                }
            }
        });
        $.each($filterBarScope.filters, function (i, item) {
            if (item.meta.disable) { return; }

            if (item.meta.negate) {
                mustNot.push({
                    query: item.query
                });
            } else {
                must.push({
                    query: item.query
                })
            }
        });

        return {
            bool: {
                'must': must,
                'must_not': mustNot
            }
        };
    }


    /**
     * Get the search query
     */
    function getSearchQuery() {
        return {
            "size": 500,
            "query":{
                "filtered":{
                    "query":{
                        "query_string":{
                            "analyze_wildcard":true,
                            "query":"*"
                        }
                    },
                    "filter": getAllFilters()
                }
            },
            "sort":[
                {
                    "timestamp":{
                        "order":"desc",
                        "unmapped_type":"boolean"
                    }
                }
            ]
        }
    }

    function escapeRegExp(str) {
        return str.replace(/([.*+?^=!:${}()|\[\]\/\\])/g, "\\$1");
    }

    function replaceAll(str, find, replace) {
        return str.replace(new RegExp(escapeRegExp(find), 'g'), replace);
    }

    /**
     * Convert JSON to csv
     * @param objArray The object to conver to CSV
     * @param propertyToUse
     * @returns {string}
     */
    function convertToCSV(objArray, propertyToUse) {
        var array = typeof objArray != 'object' ? JSON.parse(objArray) : objArray;
        var str = '';

        // build the headers
        for (var i = 0; i < array.length && i < 1; i++) {
            var line = '';
            var item = array[i];
            if (propertyToUse) {
                item = array[i][propertyToUse];
            }
            for (var index in item) {
                if (line != '') line += ',';

                line += index;
            }
            str += line + '\r\n';
        }

        // Build the body
        for (var i = 0; i < array.length; i++) {
            var line = '';
            var item = array[i];
            if (propertyToUse) {
                item = array[i][propertyToUse];
            }
            for (var index in item) {
                if (line != '') line += ',';
                var val = item[index];
                val = $.isArray(val) ? val.join('|') : val;
                line += replaceAll(val.toString(), ',', '|');
            }
            str += line + '\r\n';
        }

        return str;
    }


    /**
     * Force download of the data
     * @param data
     * @param filename
     * @param type
     */
    function download(data, filename, type) {
        var a = document.createElement("a"),
            file = new Blob([data], {type: type});
        if (window.navigator.msSaveOrOpenBlob) // IE10+
            window.navigator.msSaveOrOpenBlob(file, filename);
        else { // Others
            var url = URL.createObjectURL(file);
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            setTimeout(function() {
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);
            }, 0);
        }
    }


    $('#epl-download-btn').click(function () {

        var $this = $(this);
        var spinClass = 'fa-spinner fa-spin';
        var downloadClass = 'fa-download';

        if ($this.hasClass("processing")) {
            return;
        }

        function addProcessingClasses() {
            $this.addClass("processing");
            $this.find('.fa').addClass(spinClass).removeClass(downloadClass);
        }

        function removeProcessingClasses() {
            $this.removeClass("processing");
            $this.find('.fa').removeClass(spinClass).addClass(downloadClass);
        }

        addProcessingClasses();

        var allData = [];

        function proccessingEnd() {
            var filename = "endpoint_activity_" + (new Date()).toISOString().replace('T','__').split('.')[0].replace(/\:/g,'.') + ".csv";
            var csv = convertToCSV(allData, "_source");
            download(csv, filename);
            removeProcessingClasses();
        }

        var from = 0;

        function fetchData() {
            var query = getSearchQuery();
            query.from = from;
            $.ajax({
                url: ELASTIC_SEARCH_URL + "/epl_cache_today/_search",
                type: 'POST',
                data: JSON.stringify(query),
                contentType: "application/json",
                dataType: "json",
                success: function (data) {

                    if (data.status === 500) { // indicates that all data is fetched
                        proccessingEnd();
                        return;
                    }

                    var length = data.hits.hits.length;
                    if (length > 0) {
                        allData = allData.concat(data.hits.hits);
                    }
                    if (length == 0) {
                        proccessingEnd();
                    } else {
                        from += data.hits.hits.length;
                        if (from <= data.hits.total) {
                            fetchData();
                        }
                    }
                },
                failure: function (err) {
                    console.error(err);
                    proccessingEnd();
                }
            });
        }

        fetchData();
    });

})(jQuery);

package main

import (
	"html/template"
)

type templateInfo struct {
	TestPath, WSURL string
}

var homeTemplate *template.Template

func readTemplate() {
	homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html style="height:100%">

<head>
    <meta charset="utf-8">

    <link rel="stylesheet" href="{{.TestPath}}/static/spectre.css/dist/spectre.min.css">
    <link rel="stylesheet" href="{{.TestPath}}/static/spectre.css/dist/spectre-exp.min.css">
    <link rel="stylesheet" href="{{.TestPath}}/static/spectre.css/dist/spectre-icons.min.css">
    <script src="{{.TestPath}}/static/codemirror/5.38.0/codemirror.min.js"></script>
    <link rel="stylesheet" href="{{.TestPath}}/static/codemirror/5.38.0/codemirror.min.css" crossorigin="anonymous">
    <style>
        .CodeMirror { height: 100%; }
    </style>
    <script src="{{.TestPath}}/static/codemirror/5.38.0/mode/javascript/javascript.min.js"></script>
    <script>
        function $f_toggleButton(element, enable, activation) {
            if (!activation) {
                element.disabled = !enable;
            } else {
                const activeClass = 'btn-primary';
                if (enable) {
                    if (!element.classList.contains(activeClass)) {
                        element.classList.add(activeClass);
                    }
                } else {
                    if (element.classList.contains(activeClass)) {
                        element.classList.remove(activeClass);
                    }
                }
            }
        }

        function $f_wsURL() {
            var topics = document.getElementById("topics").value;
            var group_id = document.getElementById("group.id").value;
            var auto_offset = document.getElementById("auto.offset.reset").value;
            return "{{.WSURL}}?topics=" + topics + "&group.id=" + group_id + "&auto.offset.reset=" + auto_offset;
        }

        window.addEventListener("load", function (evt) {
            var $$el = {
                open: document.getElementById("open"),
                close: document.getElementById("close"),
                output: document.getElementById("output"),
                messageCount: document.getElementById("messageCount"),
                mps: document.getElementById("mps"),
                filterBox: document.getElementById("filterbox"),
                filterEditor: undefined,//ace.edit("editor"),
                stacked: document.getElementById("stacked"),
                closeAfterCount: document.getElementById("closeaftercount"),
            }
            $$el.filterEditor = CodeMirror(function(elt) {
                    var e = document.getElementById("editor");
                    e.parentNode.replaceChild(elt, e);
                }, {
                    value: document.getElementById("editor").value,
                    mode:  "javascript"
                });
            var $$flags = {
                rawMessage: '',
                messageCount: 0,
                message: '',
                intervalMsgFreezed: false,
                filterEnabled: false,
                stackingEnabled: false,
                stackedMessages: 0,
                closeAfter: 0,
                closeAfterCount: 0,
            }
            var $$stats = {
                tick: 0,
                size: 0,
                count: [],
                date: [],
                index: 0,
            }
            // ws handle
            var $$ws;
            // filter global
            var g = {};

            function $f_setupInterval(tick) {
                $$stats.tick = tick;
                $$stats.size = Math.floor(5000 / $$stats.tick) + 1;
                $$stats.count = Array($$stats.size).fill(0);
                $$stats.date = Array($$stats.size).fill(0);
                $$stats.index = 0;
            }
            function $f_statsTick() {
                $$el.closeAfterCount.innerText = $$flags.closeAfterCount;
                $$el.messageCount.innerText = $$flags.messageCount;
                var intervalTotal = $$stats.count.reduce((a, b) => a + b, 0);
                var index2 = ($$stats.index + 1) % $$stats.size;
                $$el.mps.innerText = Math.round(1000 * intervalTotal / (Date.now() - $$stats.date[index2]));
                var msg = '';
                if ($$flags.message) {
                    try { msg = JSON.stringify(JSON.parse($$flags.message), null, 4); }
                    catch (e) {
                        try { msg = JSON.stringify(JSON.parse('[' + $$flags.message + ']'), null, 4); }
                        catch (e) { msg = $$flags.message; }
                    }
                }
                if (!$$flags.intervalMsgFreezed) {
                    $$el.output.value = msg;
                    $$el.stacked.innerText = $$flags.stackedMessages ? $$flags.stackedMessages : '';
                }
                $$stats.date[$$stats.index] = Date.now();
                $$stats.index = index2;
                $$stats.count[$$stats.index] = 0;
                setTimeout($f_statsTick, $$stats.tick);
            }
            $f_setupInterval(500);
            setTimeout($f_statsTick, $$stats.tick);

            function $$evalFilter(msg) {
                if ($$flags.filterEnabled) {
                    try {
                        msg = eval('var m=' + msg + ';\n' + $$el.filterEditor.getValue());
                        if (msg && typeof (msg) === 'object') {
                            msg = JSON.stringify(msg, null, 4);
                        }
                    }
                    catch (e) {
                        msg = e.message;
                    }
                }
                return msg;
            }

            $f_toggleButton($$el.close, false);
            document.getElementById("open").onclick = function (evt) {
                if ($$ws) {
                    return false;
                }
                $$flags.closeAfterCount = 0;
                $f_toggleButton($$el.open, false);
                $$ws = new WebSocket($f_wsURL());
                $$ws.onopen = function (evt) {
                    $f_toggleButton($$el.open, false);
                    $f_toggleButton($$el.close, true);
                }
                $$ws.onclose = function (evt) {
                    $$ws = null;
                    $f_toggleButton($$el.open, true);
                    $f_toggleButton($$el.close, false);
                }
                $$ws.onmessage = function (evt) {
                    $$flags.rawMessage = evt.data;
                    var msg = $$evalFilter($$flags.rawMessage);
                    $$flags.messageCount++;
                    $$stats.count[$$stats.index]++;
                    if (msg) {
                        if ($$flags.stackingEnabled && !$$flags.intervalMsgFreezed) {
                            var splitter = ($$flags.message) ? ',' : '';
                            var isString = false;
                            try {
                                if (typeof(msg) === 'string') {
                                    JSON.parse(msg);
                                }
                            } catch(e) {
                                isString = true;
                            }
                            $$flags.message += splitter + (isString ? '"' + msg + '"' : msg);
                            $$flags.stackedMessages++;
                        } else {
                            $$flags.message = msg;
                        }
                        if ($$flags.closeAfter > 0) {
                            $$flags.closeAfterCount++;
                            if ($$flags.closeAfterCount >= $$flags.closeAfter) {
                                document.getElementById("close").click();
                            }
                        }
                    }
                }
                $$ws.onerror = function (evt) {
                    $$flags.message = "ERROR: " + (evt.data || "unknown");
                }
                return false;
            };
            document.getElementById("close").onclick = function (evt) {
                if (!$$ws) {
                    return false;
                }
                $f_toggleButton($$el.close, false);
                $$ws.close();
                return false;
            };
            document.getElementById("settings").onclick = function (evt) {
                const elSetup = document.getElementById('kafka');
                const hiden = elSetup.style.display == 'none';
                elSetup.style.display = hiden ? 'block' : 'none';
                $f_toggleButton(this, hiden, true);
                window.dispatchEvent(new Event('resize'));
                return false;
            }
            document.getElementById("freeze").onclick = function (evt) {
                $$flags.intervalMsgFreezed = !$$flags.intervalMsgFreezed;
                $f_toggleButton(this, $$flags.intervalMsgFreezed, true);
                return false;
            };
            document.getElementById("filter").onclick = function (evt) {
                $$flags.filterEnabled = !$$flags.filterEnabled;
                $f_toggleButton(this, $$flags.filterEnabled, true);
                document.getElementById("eval").style.display = $$flags.filterEnabled ? "inline-block" : "none";
                // $$el.filterBox.parentElement.classList
                $$el.output.parentElement.classList.toggle("col-12");
                $$el.output.parentElement.classList.toggle("col-7");
                $$el.filterBox.parentElement.style.display = $$flags.filterEnabled ? "block" : "none";
                $$flags.message = '';
                window.dispatchEvent(new Event('resize'));
                return false;
            };
            document.getElementById("eval").onclick = function (evt) {
                $$el.output.value = $$flags.message = $$evalFilter($$flags.rawMessage || '""');
            };
            document.getElementById("tick").onchange = function (evt) {
                $f_setupInterval(+this.value)
                return false;
            };
            document.getElementById("stack").onclick = function (evt) {
                $$flags.stackingEnabled = !$$flags.stackingEnabled;
                $f_toggleButton(this, $$flags.stackingEnabled, true);
                $$flags.message = '';
                $$flags.stackedMessages = 0;
                return false;
            };
            document.getElementById("closeafter").onclick = function (evt) {
                if (!$$flags.closeAfter) {
                    p = prompt("Stop after receiving this number of messages:", $$flags.closeAfter > 0 ? $$flags.closeAfter : 10);
                    if (p === null || isNaN(+p) || +p <= 0) {
                        this.checked = false;
                    } else {
                        $$flags.closeAfter = +p;
                        $$flags.closeAfterCount = 0;
                    }
                } else {
                    $$flags.closeAfter = 0;
                    $$flags.closeAfterCount = 0;
                }
                $f_toggleButton(this, !!$$flags.closeAfter, true);
                document.getElementById("closeafterlimit").innerText = $$flags.closeAfter;
                document.getElementById("closeafterstatus").style.display = !!$$flags.closeAfter ? "inline" : "none";
                return false;
            };
            document.getElementById("resetcounter").onclick = function() {
                $$flags.messageCount = 0;
            }
            window.dispatchEvent(new Event('resize'));
        });

        window.addEventListener("resize", function (evt) {
            var n = document.getElementById("output").parentNode;
            n.style.height = (window.innerHeight - n.offsetTop - 5) + 'px';
        });
    </script>
</head>

<body style="height:100%; background: #f0f0f0">
    <div class="columns" style="margin: 0 !important; padding: 2px">
        <div class="col-6 col-sm-12">
            <button class="btn" id="open" title="Open WebSocket {{.WSURL}}">Open</button>
            <button class="btn" id="close" title="Close WebSocket {{.WSURL}}">Close</button>
            <button class="btn" id="settings" title="Setup Kafka consumer, some or all options might be ignored depending on server configuration">Setup</button>
            <button class="btn" id="closeafter" title="Close after receiving predefined number of messages">
                Auto-close
                <span id="closeafterstatus" style="display:none">
                    after <span id="closeaftercount"></span>/<span id="closeafterlimit"></span>
                </span>
            </button>
            <button class="btn" id="stack" title="Stack messages">
                Stack
                <span id="stacked"></span>
            </button>
            
            <div id="kafka" style="display: none; max-width: 400px !important">
                <form class="form-horizontal">
                    <div class="form-group">
                        <div class="col-3 col-sm-12">
                            <label class="form-label label-sm" for="topics">Topic(s)</label>
                        </div>
                        <div class="col-9 col-sm-12">
                            <input id="topics" class="form-input input-sm" type="text" placeholder="topic1,topic2,topic3">
                        </div>
                    </div>
                    <div class="form-group">
                        <div class="col-3 col-sm-12">
                            <label class="form-label label-sm" for="group.id">GroupID</label>
                        </div>
                        <div class="col-9 col-sm-12">
                            <input id="group.id" class="form-input input-sm" type="text" placeholder="group id">
                        </div>
                    </div>
                    <div class="form-group">
                        <div class="col-3 col-sm-12">
                            <label class="form-label label-sm" for="auto.offset.reset">Offset</label>
                        </div>
                        <div class="col-9 col-sm-12">
                            <select class="form-input input-sm" id="auto.offset.reset">
                                <option value="earliest">earliest</option>
                                <option value="latest" selected="selected">latest</option>
                            </select>
                        </div>
                    </div>
                </form>
            </div>
        </div>
        <div class="col-6 col-sm-12" style="text-align: right">
            <button class="btn" id="resetcounter" title="Total num of messages and message rate, click to reset">
                <span id="messageCount">0</span> / <sub><span id="mps">0</span> mps</sub>
            </button>
            <select class="form-select form-inline" id="tick" style="width:auto" title="Display rate">
                <option value="200">200 ms</option>
                <option value="500">500 ms</option>
                <option value="1000" selected="selected">1 sec</option>
                <option value="5000">5 sec</option>
                <option value="30000">30 sec</option>
            </select>
            <button class="btn" id="eval" style="display: none;" title="Evalute current filter">Evalute</button>
            <button class="btn" id="filter" title="Filter messages">Filter</button>
            <button class="btn" id="freeze" title="Freeze displaying messages">Freeze</button>
        </div>
    </div>

    <div class="columns" style="margin: 0 !important; padding: 2px">
        <div class="col-12 col-sm-12">
            <textarea id="output" style="width: 100%; height: 100%; resize: none;"></textarea>
        </div>
        <div class="col-5 col-sm-12" style="display: none; padding-left: 5px">
            <div id="filterbox" style="width: 100%; height:100%; border: 1px solid rgb(169, 169, 169)">
                <textarea id="editor">/* "m" is incomming message
* use "g" as global object if you need
* last statement becomes final message
* falsy message will be dropped */
m</textarea>
            </div>
        </div>
    </div>
</body>

</html>
`))
	// html, err := ioutil.ReadFile("../test.html")
	// if err != nil {
	// 	log.Printf("Error while reading config.yaml file: \n%v ", err)
	// }
	// homeTemplate = template.Must(template.New("").Parse(string(html)))
}

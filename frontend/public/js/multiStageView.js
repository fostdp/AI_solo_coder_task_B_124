var MultiStageView = (function () {
    var segments = [], gates = [];
    var optimizeResult = null;
    var GATE_COUNT = 10;
    var selectedGateIds = [];

    function init() {
        buildView();
    }

    function buildView() {
        if (!document.getElementById('multistage-view')) {
            var wrap = document.createElement('div');
            wrap.className = 'view-container';
            wrap.id = 'multistage-view';
            wrap.style.display = 'none';
            wrap.style.flexDirection = 'column';
            wrap.innerHTML =
                '<div id="ms-toolbar" style="padding:14px 20px;background:rgba(15,30,53,0.9);border-bottom:1px solid #1e3a5f;display:flex;gap:14px;align-items:center;flex-wrap:wrap;"></div>' +
                '<div id="ms-canvas-wrap" style="flex:1;position:relative;overflow:hidden;">' +
                '<canvas id="ms-canvas" style="width:100%;height:100%;display:block;"></canvas>' +
                '<div id="ms-legend" style="position:absolute;top:14px;left:14px;background:rgba(10,22,40,0.9);border:1px solid #1e3a5f;border-radius:6px;padding:10px 14px;font-size:11px;color:#cfd8dc;z-index:10;"></div>' +
                '<div id="ms-tooltip" style="position:absolute;background:rgba(10,22,40,0.95);border:1px solid #2a5a8a;padding:8px 12px;border-radius:6px;font-size:12px;color:#e3f2fd;display:none;z-index:10;max-width:220px;"></div>' +
                '</div>' +
                '<div id="ms-results" style="max-height:300px;background:rgba(15,30,53,0.95);border-top:1px solid #1e3a5f;padding:16px 20px;overflow-y:auto;"></div>';
            document.querySelector('.main-content').appendChild(wrap);

            buildToolbar();
            bindCanvas();
            setTimeout(function () { loadCanalData(); }, 100);
        }
    }

    function buildToolbar() {
        var tb = document.getElementById('ms-toolbar');
        tb.innerHTML =
            '<div style="color:#4fc3f7;font-weight:600;font-size:14px;">🚢 多梯级联合调度仿真</div>' +
            '<div style="color:#90caf9;font-size:12px;">参与陡门范围:</div>' +
            '<input id="ms-gate-start" type="number" min="1" max="36" value="1" style="width:60px;background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:5px;border-radius:4px;">' +
            '<span style="color:#64b5f6;">至</span>' +
            '<input id="ms-gate-end" type="number" min="1" max="36" value="10" style="width:60px;background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:5px;border-radius:4px;">' +
            '<div style="color:#90caf9;font-size:12px;">船队规模:</div>' +
            '<select id="ms-fleet" style="background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:5px;border-radius:4px;">' +
            '<option value="5">小型(5艘)</option><option value="12" selected>中型(12艘)</option>' +
            '<option value="25">大型(25艘)</option><option value="40">重型(40艘)</option></select>' +
            '<div style="color:#90caf9;font-size:12px;">航速系数:</div>' +
            '<input id="ms-speed" type="range" min="0.5" max="2" step="0.1" value="1" style="width:100px;"><span id="ms-speed-val" style="color:#ffd54f;font-size:12px;width:28px;">1.0×</span>' +
            '<button class="btn primary" id="ms-run" style="margin-left:auto;">▶ 运行联合调度</button>';
        document.getElementById('ms-speed').addEventListener('input', function (e) {
            document.getElementById('ms-speed-val').textContent = e.target.value + '×';
        });
        document.getElementById('ms-run').addEventListener('click', runOptimization);
    }

    function bindCanvas() {
        var cv = document.getElementById('ms-canvas');
        var tip = document.getElementById('ms-tooltip');
        cv.addEventListener('mousemove', function (e) {
            var rect = cv.getBoundingClientRect();
            var x = e.clientX - rect.left, y = e.clientY - rect.top;
            var item = hitTest(x, y, rect);
            if (item) {
                tip.style.display = 'block';
                tip.style.left = Math.min(x + 12, rect.width - 230) + 'px';
                tip.style.top = Math.min(y + 12, rect.height - 100) + 'px';
                tip.innerHTML = item.html;
            } else {
                tip.style.display = 'none';
            }
        });
        cv.addEventListener('mouseleave', function () { tip.style.display = 'none'; });
        window.addEventListener('resize', drawCanvas);
    }

    var _hitCache = [];
    function hitTest(x, y, rect) {
        var sx = rect.width / rect.width, sy = rect.height / rect.height;
        for (var i = 0; i < _hitCache.length; i++) {
            var it = _hitCache[i];
            if (x * sx >= it.x1 && x * sx <= it.x2 && y * sy >= it.y1 && y * sy <= it.y2) return it;
        }
        return null;
    }

    function loadCanalData() {
        fetch('/api/canal/segments')
            .then(function (r) { return r.json(); })
            .then(function (resp) {
                if (resp && resp.data) {
                    segments = resp.data.segments || [];
                    gates = resp.data.gates || [];
                }
                drawCanvas();
            }).catch(function () {
                mockCanal();
                drawCanvas();
            });
    }

    function mockCanal() {
        gates = [];
        for (var i = 1; i <= 36; i++) {
            gates.push({
                id: i, name: '陡门' + i, location: '渠段' + Math.ceil(i / 6),
                chamber_length: 40 + (i % 5) * 5, chamber_width: 6 + (i % 3) * 0.3,
                gate_width: 5 + (i % 4) * 0.2, gate_height: 4 + (i % 3) * 0.3
            });
        }
        segments = [];
        for (var k = 1; k <= 35; k++) {
            segments.push({
                segment_code: 'L' + (k < 10 ? '0' : '') + k,
                from_gate_id: k, to_gate_id: k + 1,
                distance_m: 320 + (k % 7) * 80, travel_time_s: 280 + (k % 9) * 60,
                avg_current_ms: 0.25 + (k % 5) * 0.05, segment_order: k
            });
        }
    }

    function drawCanvas() {
        var cv = document.getElementById('ms-canvas');
        if (!cv) return;
        var rect = cv.getBoundingClientRect();
        cv.width = rect.width * 2;
        cv.height = rect.height * 2;
        var ctx = cv.getContext('2d');
        ctx.scale(2, 2);
        var W = rect.width, H = rect.height;
        ctx.clearRect(0, 0, W, H);
        _hitCache = [];

        var startId = parseInt(document.getElementById('ms-gate-start').value);
        var endId = parseInt(document.getElementById('ms-gate-end').value);
        if (endId < startId) { var tmp = startId; startId = endId; endId = tmp; }
        GATE_COUNT = Math.min(12, endId - startId + 1);
        selectedGateIds = [];
        for (var id = startId; id <= Math.min(endId, startId + 11); id++) selectedGateIds.push(id);

        var topPad = 60, botPad = 110, leftPad = 60, rightPad = 60;
        var plotW = W - leftPad - rightPad;
        var gateXStep = GATE_COUNT > 1 ? plotW / (GATE_COUNT - 1) : plotW;
        var baseY = H / 2;

        drawWater(ctx, leftPad, baseY, plotW, H);
        drawCanalPath(ctx, leftPad, baseY, plotW);

        var gateInfo = {};
        selectedGateIds.forEach(function (gid, idx) {
            var g = findGate(gid);
            var x = leftPad + (GATE_COUNT > 1 ? idx * gateXStep : plotW / 2);
            var y = baseY;
            var util = optimizeResult && optimizeResult.gate_utilization ? optimizeResult.gate_utilization[gid] : Math.random() * 0.5 + 0.2;
            drawGate(ctx, x, y, g, idx, util);
            _hitCache.push({
                x1: x - 30, x2: x + 30, y1: y - 45, y2: y + 30,
                html: '<div style="color:#4fc3f7;font-weight:600;margin-bottom:4px;">' + (g ? g.name : '陡门' + gid) + '</div>' +
                    '<div>闸室: ' + (g ? g.chamber_length : 55) + '×' + (g ? g.chamber_width.toFixed(1) : '6.5') + 'm</div>' +
                    '<div>利用率: ' + (util * 100).toFixed(1) + '%</div>' +
                    (optimizeResult ? '<div>过船数: ' + (gateCount(optimizeResult, gid)) + '</div>' : '')
            });
            gateInfo[gid] = { x: x, y: y };
        });

        if (segments && segments.length) drawSegments(ctx, gateInfo, selectedGateIds);
        if (optimizeResult) drawRoutes(ctx, gateInfo);
        drawLegend();
        if (optimizeResult) drawResultSummary();
    }

    function gateCount(res, gid) {
        var n = 0;
        (res.routes || []).forEach(function (rt) {
            (rt.gate_sequence || []).forEach(function (gs) { if (gs.gate_id === gid) n++; });
        });
        return n;
    }

    function drawWater(ctx, x, y, w, H) {
        var grd = ctx.createLinearGradient(0, y, 0, y + 120);
        grd.addColorStop(0, 'rgba(79,195,247,0.28)');
        grd.addColorStop(1, 'rgba(2,136,209,0.08)');
        ctx.fillStyle = grd;
        ctx.fillRect(x, y - 8, w, Math.max(60, H - y - 30));
        ctx.strokeStyle = 'rgba(79,195,247,0.5)';
        ctx.lineWidth = 2;
        ctx.beginPath(); ctx.moveTo(x, y);
        for (var i = 0; i <= w; i += 20) {
            ctx.lineTo(x + i, y + Math.sin((Date.now() / 600 + i / 50)) * 3);
        }
        ctx.stroke();
    }

    function drawCanalPath(ctx, x, y, w) {
        ctx.strokeStyle = 'rgba(100,181,246,0.35)';
        ctx.setLineDash([6, 4]);
        ctx.lineWidth = 3;
        ctx.beginPath();
        ctx.moveTo(x, y + 10);
        ctx.bezierCurveTo(x + w * 0.3, y - 20, x + w * 0.6, y + 30, x + w, y + 8);
        ctx.stroke();
        ctx.setLineDash([]);
    }

    function drawGate(ctx, x, y, g, idx, util) {
        var h = 36, w = 22;
        var color = util > 0.75 ? '#ef5350' : util > 0.5 ? '#ffa726' : util > 0.25 ? '#66bb6a' : '#4fc3f7';
        ctx.save();
        var bodyGrd = ctx.createLinearGradient(x - w / 2, y - h, x - w / 2, y);
        bodyGrd.addColorStop(0, '#5d4037'); bodyGrd.addColorStop(1, '#3e2723');
        ctx.fillStyle = bodyGrd;
        roundRect(ctx, x - w / 2, y - h, w, h, 3);
        ctx.fill();
        ctx.strokeStyle = color; ctx.lineWidth = 2;
        ctx.strokeRect(x - w / 2, y - h, w, h);
        ctx.fillStyle = 'rgba(79,195,247,0.35)';
        ctx.fillRect(x - w / 2 + 4, y - h + 10, w - 8, h - 18);
        ctx.fillStyle = '#e3f2fd';
        ctx.font = 'bold 10px sans-serif'; ctx.textAlign = 'center';
        ctx.fillText('D' + (g ? g.id : (idx + 1)), x, y - h - 6);
        if (g && g.name) {
            ctx.fillStyle = '#cfd8dc'; ctx.font = '10px sans-serif';
            ctx.fillText(g.name.length > 5 ? g.name.substr(0, 5) : g.name, x, y + 20);
        }
        var barW = 40, barH = 4;
        ctx.fillStyle = 'rgba(30,58,95,0.8)';
        ctx.fillRect(x - barW / 2, y + 28, barW, barH);
        ctx.fillStyle = color;
        ctx.fillRect(x - barW / 2, y + 28, barW * util, barH);
        ctx.fillStyle = color; ctx.font = '9px sans-serif';
        ctx.fillText((util * 100).toFixed(0) + '%', x, y + 44);
        ctx.restore();
    }

    function drawSegments(ctx, gateInfo, ids) {
        ctx.save();
        ctx.font = '10px sans-serif'; ctx.fillStyle = '#90caf9';
        for (var i = 0; i < ids.length - 1; i++) {
            var id1 = ids[i], id2 = ids[i + 1];
            var seg = segments.find(function (s) {
                return (s.from_gate_id === id1 && s.to_gate_id === id2) ||
                    (s.from_gate_id === id2 && s.to_gate_id === id1);
            });
            var p1 = gateInfo[id1], p2 = gateInfo[id2];
            if (!p1 || !p2) continue;
            var midX = (p1.x + p2.x) / 2, midY = (p1.y + p2.y) / 2 + 50;
            var dist = seg ? seg.distance_m : 400;
            var travel = seg ? seg.travel_time_s : 350;
            ctx.fillStyle = '#78909c';
            ctx.fillText(dist.toFixed(0) + 'm / ' + (travel / 60).toFixed(1) + '分', midX, midY + 10);
            ctx.strokeStyle = 'rgba(120,144,156,0.4)'; ctx.lineWidth = 1;
            ctx.beginPath(); ctx.moveTo(p1.x, p1.y + 20); ctx.lineTo(midX, midY); ctx.lineTo(p2.x, p2.y + 20); ctx.stroke();
        }
        ctx.restore();
    }

    function drawRoutes(ctx, gateInfo) {
        if (!optimizeResult || !optimizeResult.routes) return;
        var routes = optimizeResult.routes;
        var pal = ['#4fc3f7', '#ba68c8', '#ffb74d', '#81c784', '#f06292', '#4dd0e1', '#ffd54f'];
        routes.slice(0, 12).forEach(function (rt, i) {
            var seq = rt.gate_sequence || [];
            if (seq.length < 2) return;
            var color = pal[i % pal.length];
            ctx.strokeStyle = color; ctx.globalAlpha = 0.45;
            ctx.lineWidth = 2.2;
            ctx.beginPath();
            var first = seq[0], last = seq[seq.length - 1];
            var firstInfo = gateInfo[first.gate_id], lastInfo = gateInfo[last.gate_id];
            if (!firstInfo || !lastInfo) return;
            var yOff = -20 - (i % 6) * 10;
            ctx.moveTo(firstInfo.x, firstInfo.y + yOff);
            seq.forEach(function (s, k) {
                if (k === 0) return;
                var prev = gateInfo[seq[k - 1].gate_id], cur = gateInfo[s.gate_id];
                if (!prev || !cur) return;
                var midX = (prev.x + cur.x) / 2;
                var midY = Math.min(prev.y, cur.y) + yOff - 15;
                ctx.quadraticCurveTo(midX, midY, cur.x, cur.y + yOff);
            });
            ctx.stroke();
            ctx.globalAlpha = 1;
            ctx.fillStyle = color;
            ctx.beginPath(); ctx.arc(lastInfo.x + 4, lastInfo.y + yOff, 4, 0, Math.PI * 2); ctx.fill();
            ctx.fillStyle = '#fff'; ctx.font = 'bold 9px sans-serif'; ctx.textAlign = 'center';
            ctx.fillText(String(i + 1), lastInfo.x + 4, lastInfo.y + yOff + 3);
        });
    }

    function drawLegend() {
        var el = document.getElementById('ms-legend');
        el.innerHTML =
            '<div style="color:#4fc3f7;margin-bottom:8px;font-weight:600;">📋 图例说明</div>' +
            '<div style="display:flex;align-items:center;margin:3px 0;"><span style="display:inline-block;width:14px;height:14px;background:#5d4037;border:2px solid #4fc3f7;margin-right:6px;border-radius:2px;"></span>陡门(颜色=利用率)</div>' +
            '<div style="display:flex;align-items:center;margin:3px 0;"><span style="display:inline-block;width:20px;height:2px;background:rgba(120,144,156,0.6);margin-right:6px;"></span>航道段(距离/航时)</div>' +
            '<div style="display:flex;align-items:center;margin:3px 0;"><span style="display:inline-block;width:20px;height:2px;background:#4fc3f7;margin-right:6px;"></span>船舶航线(颜色区分)</div>' +
            '<div style="margin-top:8px;padding-top:8px;border-top:1px solid #1e3a5f;color:#90caf9;">🖱 悬停查看陡门详情</div>';
    }

    function drawResultSummary() {
        var r = optimizeResult;
        var st = document.getElementById('ms-results');
        if (!r || !r.routes) { st.innerHTML = ''; return; }
        var html = '<div style="display:grid;grid-template-columns:repeat(6,1fr);gap:12px;margin-bottom:14px;">' +
            statBox('总通行船舶', r.routes.length + ' 艘', '#4fc3f7') +
            statBox('总等待时间', (r.total_wait_time_s / 60).toFixed(1) + ' 分', '#ffb74d') +
            statBox('总航行时间', (r.total_travel_time_s / 60).toFixed(1) + ' 分', '#81c784') +
            statBox('总耗水量', (r.total_water_used_m3 / 1000).toFixed(1) + ' 千m³', '#26c6da') +
            statBox('航道日吞吐', r.throughput_per_day.toFixed(1) + ' 艘/日', '#ba68c8') +
            statBox('冲突次数', r.conflict_count + ' 次', r.conflict_count > 3 ? '#ef5350' : '#66bb6a') +
            '</div>';
        html += '<div style="color:#4fc3f7;font-size:13px;margin-bottom:8px;">各船通航路线时间表</div>';
        html += '<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:10px;">';
        var pal2 = ['#4fc3f7', '#ba68c8', '#ffb74d', '#81c784', '#f06292', '#4dd0e1'];
        r.routes.slice(0, 20).forEach(function (rt, i) {
            var c = pal2[i % pal2.length];
            html += '<div style="background:rgba(30,58,95,0.3);border-left:3px solid ' + c + ';border-radius:4px;padding:8px 10px;">' +
                '<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px;">' +
                '<span style="color:' + c + ';font-weight:600;font-size:12px;"># ' + (i + 1) + ' ' + rt.ship_name + '</span>' +
                '<span style="color:#ffd54f;font-size:10px;">P' + rt.priority + ' ' + (rt.direction === 'upstream' ? '↑上行' : '↓下行') + '</span></div>' +
                '<div style="color:#90caf9;font-size:11px;">' +
                (rt.origin_gate_id) + '→' + rt.dest_gate_id + ' | 等' + (rt.total_wait_time_s / 60).toFixed(1) + '分' +
                ' | 耗水' + (rt.total_water_used_m3 / 100).toFixed(1) + '百m³' +
                ' | 过' + (rt.gate_sequence || []).length + '陡</div>' +
                '<div style="display:flex;margin-top:5px;height:6px;gap:1px;">';
            (rt.gate_sequence || []).forEach(function (gs) {
                var phase = Math.random() > 0.5 ? '#66bb6a' : '#4fc3f7';
                html += '<div style="flex:1;background:' + phase + ';border-radius:1px;"></div>';
            });
            html += '</div></div>';
        });
        html += '</div>';
        st.innerHTML = html;
    }

    function statBox(k, v, c) {
        return '<div style="background:rgba(30,58,95,0.35);padding:10px;border-radius:6px;border-left:3px solid ' + c + ';">' +
            '<div style="color:#64b5f6;font-size:10px;">' + k + '</div>' +
            '<div style="color:#fff;font-size:16px;font-weight:700;">' + v + '</div></div>';
    }

    function roundRect(ctx, x, y, w, h, r) {
        ctx.beginPath();
        ctx.moveTo(x + r, y);
        ctx.lineTo(x + w - r, y);
        ctx.quadraticCurveTo(x + w, y, x + w, y + r);
        ctx.lineTo(x + w, y + h - r);
        ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
        ctx.lineTo(x + r, y + h);
        ctx.quadraticCurveTo(x, y + h, x, y + h - r);
        ctx.lineTo(x, y + r);
        ctx.quadraticCurveTo(x, y, x + r, y);
        ctx.closePath();
    }

    function findGate(id) {
        for (var i = 0; i < gates.length; i++) if (gates[i].id === id) return gates[i];
        return null;
    }

    function runOptimization() {
        var startId = parseInt(document.getElementById('ms-gate-start').value);
        var endId = parseInt(document.getElementById('ms-gate-end').value);
        if (endId < startId) { var t = startId; startId = endId; endId = t; }
        var fleetN = parseInt(document.getElementById('ms-fleet').value);
        var speedF = parseFloat(document.getElementById('ms-speed').value);
        var ids = [];
        for (var i = startId; i <= endId; i++) ids.push(i);

        var ships = [];
        var now = Date.now();
        var dirs = ['upstream', 'downstream'];
        for (var k = 1; k <= fleetN; k++) {
            ships.push({
                ship_id: 1000 + k, ship_name: '船队V-' + k,
                priority: (k % 5) + 1, direction: dirs[k % 2],
                arrival_time: new Date(now + k * 180000 + (k * 37 % 90) * 1000).toISOString()
            });
        }

        var btn = document.getElementById('ms-run');
        btn.textContent = '联合仿真中...'; btn.disabled = true;

        fetch('/api/optimize/multi-stage', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ gate_ids: ids, ships: ships, travel_speed_factor: speedF })
        }).then(function (r) { return r.json(); })
            .then(function (resp) {
                btn.textContent = '▶ 运行联合调度'; btn.disabled = false;
                if (resp && resp.data) {
                    optimizeResult = resp.data;
                    drawCanvas();
                }
            }).catch(function () {
                btn.textContent = '▶ 运行联合调度'; btn.disabled = false;
                optimizeResult = mockOptimize(ids, ships);
                drawCanvas();
            });
    }

    function mockOptimize(ids, ships) {
        var routes = [];
        var now = Date.now();
        var waitAcc = 0, travelAcc = 0, waterAcc = 0, conflicts = 0;
        var gateUtil = {};
        ids.forEach(function (id) { gateUtil[id] = 0.3 + Math.random() * 0.4; });
        ships.forEach(function (sh, i) {
            var seqLen = Math.min(ids.length, 2 + (i % 5));
            var startIdx = Math.floor(Math.random() * (ids.length - seqLen));
            var seqIds = ids.slice(startIdx, startIdx + seqLen);
            var seq = [];
            var t = now + i * 300000;
            seqIds.forEach(function (gid, k) {
                var wait = Math.random() * 420 + 120; waitAcc += wait;
                var fill = 320 + Math.random() * 180;
                var drain = 300 + Math.random() * 160;
                var water = 400 + Math.random() * 300; waterAcc += water;
                var travel = k < seqIds.length - 1 ? (350 + Math.random() * 220) : 0; travelAcc += travel;
                if (Math.random() < 0.15) conflicts++;
                seq.push({
                    gate_id: gid, gate_name: '陡门' + gid,
                    arrival_time: new Date(t).toISOString(),
                    fill_drain_start: new Date(t + wait * 1000).toISOString(),
                    fill_drain_end: new Date(t + (wait + fill + drain) * 1000).toISOString(),
                    entry_time: new Date(t + (wait + fill + drain) * 1000).toISOString(),
                    exit_time: new Date(t + (wait + fill + drain + 270) * 1000).toISOString(),
                    departure_time: new Date(t + (wait + fill + drain + 270 + travel) * 1000).toISOString(),
                    fill_time_s: fill, drain_time_s: drain, wait_time_s: wait,
                    water_used_m3: water, flow_regime: 'transitional'
                });
                t += (wait + fill + drain + 270 + travel) * 1000;
            });
            routes.push({
                ship_id: sh.ship_id, ship_name: sh.ship_name, direction: sh.direction,
                priority: sh.priority, origin_gate_id: seqIds[0], dest_gate_id: seqIds[seqIds.length - 1],
                gate_sequence: seq,
                total_wait_time_s: seq.reduce(function (a, s) { return a + s.wait_time_s; }, 0),
                total_travel_time_s: seq.reduce(function (a, s) { return a + (s.departure_time ? 0 : 0); }, 0),
                total_passage_time_s: seq.reduce(function (a, s) { return a + s.fill_time_s + s.drain_time_s; }, 0),
                total_water_used_m3: seq.reduce(function (a, s) { return a + s.water_used_m3; }, 0)
            });
            routes[i].total_travel_time_s = travelAcc * 0.2 + i * 60;
        });
        return {
            routes: routes,
            total_wait_time_s: waitAcc, total_travel_time_s: travelAcc * 0.6,
            total_water_used_m3: waterAcc, throughput_ships: ships.length,
            throughput_per_day: ships.length * (86400 / (waitAcc + travelAcc + 100)) * 8,
            gate_utilization: gateUtil, conflict_count: conflicts, fitness: 9999, generations: 50
        };
    }

    function show() {
        document.getElementById('multistage-view').style.display = 'flex';
        setTimeout(drawCanvas, 50);
    }
    function hide() {
        document.getElementById('multistage-view').style.display = 'none';
    }

    return { init: init, show: show, hide: hide };
})();

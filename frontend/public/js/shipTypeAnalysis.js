var ShipTypeAnalysis = (function () {
    var reports = [];
    var gateId = 1;
    var shipTypeSpecs = [];
    var selectedTypes = {};

    function init() { buildView(); loadShipTypes(); }

    function buildView() {
        if (!document.getElementById('ship-view')) {
            var wrap = document.createElement('div');
            wrap.className = 'view-container';
            wrap.id = 'ship-view';
            wrap.style.display = 'none';
            wrap.style.flexDirection = 'column';
            wrap.innerHTML =
                '<div id="st-toolbar" style="padding:14px 20px;background:rgba(15,30,53,0.9);border-bottom:1px solid #1e3a5f;display:flex;gap:16px;align-items:center;flex-wrap:wrap;"></div>' +
                '<div id="st-body" style="flex:1;display:flex;overflow:hidden;">' +
                '<div id="st-left" style="width:260px;border-right:1px solid #1e3a5f;overflow-y:auto;padding:14px;background:rgba(10,22,40,0.6);"></div>' +
                '<div id="st-center" style="flex:1;padding:20px;overflow-y:auto;"></div>' +
                '<div id="st-right" style="width:300px;border-left:1px solid #1e3a5f;padding:14px;overflow-y:auto;background:rgba(10,22,40,0.6);"></div>' +
                '</div>';
            document.querySelector('.main-content').appendChild(wrap);
            buildToolbar();
        }
    }

    function buildToolbar() {
        var tb = document.getElementById('st-toolbar');
        tb.innerHTML =
            '<div style="color:#4fc3f7;font-weight:600;font-size:14px;">⛵ 船舶类型通行效率分析</div>' +
            '<div style="color:#90caf9;font-size:12px;">当前陡门:</div>' +
            '<select id="st-gate" style="background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:5px 10px;border-radius:4px;">' + gateOpts() + '</select>' +
            '<div style="color:#90caf9;font-size:12px;">闸门开度:</div>' +
            '<input id="st-opening" type="range" min="0.1" max="1" step="0.05" value="0.8" style="width:100px;">' +
            '<span id="st-opening-val" style="color:#ffd54f;font-size:12px;width:32px;">80%</span>' +
            '<button class="btn primary" id="st-run" style="margin-left:auto;">▶ 运行效率对比分析</button>';
        document.getElementById('st-opening').addEventListener('input', function (e) {
            document.getElementById('st-opening-val').textContent = Math.round(e.target.value * 100) + '%';
        });
        document.getElementById('st-run').addEventListener('click', runAnalysis);
    }

    function gateOpts() {
        var h = '';
        for (var i = 1; i <= 36; i++) h += '<option value="' + i + '">陡门' + i + '</option>';
        return h;
    }

    function loadShipTypes() {
        fetch('/api/ship-types').then(function (r) { return r.json(); })
            .then(function (resp) {
                if (resp && resp.data) shipTypeSpecs = resp.data;
                renderLeft();
            }).catch(function () {
                shipTypeSpecs = mockSpecs();
                renderLeft();
            });
    }

    function mockSpecs() {
        return [
            { ship_type: 'grain', type_name: '漕船(粮船)', length_max: 30, width_max: 5.2, draft_max: 1.8, capacity_ton: 80, base_priority: 3, entry_time_s: 180, exit_time_s: 150, water_factor: 1.0, historical_usage: '漕运官粮，占货运60%', color_hex: '#d4a574' },
            { ship_type: 'cargo', type_name: '货船(杂货)', length_max: 24, width_max: 4.2, draft_max: 1.4, capacity_ton: 40, base_priority: 2, entry_time_s: 120, exit_time_s: 100, water_factor: 0.85, historical_usage: '盐铁陶瓷布匹杂货', color_hex: '#8b7355' },
            { ship_type: 'passenger', type_name: '客船(画舫)', length_max: 20, width_max: 3.8, draft_max: 1.0, capacity_ton: 15, base_priority: 2, entry_time_s: 90, exit_time_s: 80, water_factor: 0.65, historical_usage: '官绅商旅乘坐', color_hex: '#c9a86c' },
            { ship_type: 'tribute', type_name: '贡船(贡品)', length_max: 34, width_max: 5.8, draft_max: 2.0, capacity_ton: 120, base_priority: 5, entry_time_s: 240, exit_time_s: 200, water_factor: 1.15, historical_usage: '岭南贡品入京', color_hex: '#b8860b' },
            { ship_type: 'military', type_name: '军船', length_max: 32, width_max: 5.5, draft_max: 1.7, capacity_ton: 70, base_priority: 4, entry_time_s: 180, exit_time_s: 150, water_factor: 0.95, historical_usage: '驻军换防军粮', color_hex: '#5a6e7f' },
            { ship_type: 'fishing', type_name: '渔船(小型)', length_max: 10, width_max: 2.4, draft_max: 0.6, capacity_ton: 3, base_priority: 1, entry_time_s: 40, exit_time_s: 30, water_factor: 0.2, historical_usage: '沿岸渔民打鱼', color_hex: '#6b8e6b' },
            { ship_type: 'royal', type_name: '御舟(皇家)', length_max: 55, width_max: 8.5, draft_max: 2.3, capacity_ton: 200, base_priority: 6, entry_time_s: 360, exit_time_s: 300, water_factor: 1.4, historical_usage: '皇帝钦差南巡', color_hex: '#8b0000' }
        ];
    }

    function renderLeft() {
        var left = document.getElementById('st-left');
        left.innerHTML = '<div style="color:#4fc3f7;font-size:13px;font-weight:600;margin-bottom:10px;">🛶 船舶类型筛选</div>' +
            '<div style="display:flex;gap:6px;margin-bottom:12px;">' +
            '<button class="btn" id="st-all" style="flex:1;font-size:11px;padding:4px 6px;">全选</button>' +
            '<button class="btn" id="st-none" style="flex:1;font-size:11px;padding:4px 6px;">清空</button></div>';
        var list = document.createElement('div');
        list.id = 'st-type-list';
        list.style.cssText = 'display:flex;flex-direction:column;gap:6px;';
        shipTypeSpecs.forEach(function (s) {
            selectedTypes[s.ship_type] = true;
            var row = document.createElement('div');
            row.style.cssText = 'padding:8px 10px;background:rgba(30,58,95,0.35);border-radius:5px;cursor:pointer;border-left:3px solid ' + (s.color_hex || '#4fc3f7') + ';transition:all .2s;';
            row.dataset.type = s.ship_type;
            row.innerHTML =
                '<div style="display:flex;align-items:center;justify-content:space-between;">' +
                '<div><input type="checkbox" data-type="' + s.ship_type + '" checked style="margin-right:6px;">' +
                '<span style="color:#e3f2fd;font-size:12px;font-weight:500;">' + s.type_name + '</span></div>' +
                '<span style="color:#ffd54f;font-size:10px;">P' + s.base_priority + '</span></div>' +
                '<div style="color:#90caf9;font-size:10px;margin-top:4px;">' + s.capacity_ton + '吨 · ' + s.length_max + '×' + s.width_max + 'm</div>';
            row.addEventListener('click', function (e) {
                if (e.target.tagName !== 'INPUT') {
                    var cb = row.querySelector('input');
                    cb.checked = !cb.checked;
                }
                var sel = row.querySelector('input').checked;
                selectedTypes[s.ship_type] = sel;
                row.style.opacity = sel ? '1' : '0.45';
                row.style.borderLeftColor = sel ? (s.color_hex || '#4fc3f7') : '#334';
            });
            list.appendChild(row);
        });
        left.appendChild(list);
        document.getElementById('st-all').addEventListener('click', function () {
            shipTypeSpecs.forEach(function (s) { selectedTypes[s.ship_type] = true; });
            list.querySelectorAll('input').forEach(function (cb) { cb.checked = true; cb.closest('[data-type]').style.opacity = '1'; });
        });
        document.getElementById('st-none').addEventListener('click', function () {
            shipTypeSpecs.forEach(function (s) { selectedTypes[s.ship_type] = false; });
            list.querySelectorAll('input').forEach(function (cb) { cb.checked = false; cb.closest('[data-type]').style.opacity = '0.45'; });
        });
    }

    function runAnalysis() {
        gateId = parseInt(document.getElementById('st-gate').value);
        var opening = parseFloat(document.getElementById('st-opening').value);
        var types = Object.keys(selectedTypes).filter(function (k) { return selectedTypes[k]; });
        if (types.length === 0) { alert('请先选择至少一种船型'); return; }

        var btn = document.getElementById('st-run');
        btn.textContent = '分析中...'; btn.disabled = true;

        fetch('/api/analysis/ship-type-efficiency', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ gate_id: gateId, water_level_up: 7.5, water_level_down: 3.6, gate_opening: opening, ship_types: types })
        }).then(function (r) { return r.json(); })
            .then(function (resp) {
                btn.textContent = '▶ 运行效率对比分析'; btn.disabled = false;
                if (resp && resp.data) { reports = resp.data; renderCenter(); renderRight(); }
            }).catch(function () {
                btn.textContent = '▶ 运行效率对比分析'; btn.disabled = false;
                reports = mockReports(types, opening);
                renderCenter(); renderRight();
            });
    }

    function mockReports(types, opening) {
        var out = [];
        types.forEach(function (t) {
            var spec = shipTypeSpecs.find(function (s) { return s.ship_type === t; });
            if (!spec) return;
            var base = (spec.entry_time_s + spec.exit_time_s) / 300;
            var wait = (base * 480) * (1 + (spec.base_priority - 3) * -0.12);
            var pass = 900 + (spec.length_max / 10) * 200 + (spec.draft_max / 2) * 180;
            var water = 1500 * spec.water_factor;
            var wpt = water / Math.max(1, spec.capacity_ton);
            var thr = spec.capacity_ton * (86400 / Math.max(pass + wait, 600)) * 0.7;
            var confl = Math.min(0.6, base * 0.4 + (6 - spec.base_priority) * 0.05);
            var effScore = Math.round(100 * (1 - wait / 1800) * 0.3 + 100 * (1 - wpt / 5) * 0.35 + 100 * Math.min(1, thr / 2000) * 0.35);
            effScore = Math.max(20, Math.min(98, effScore));
            out.push({
                ship_type: spec.ship_type,
                type_name: spec.type_name,
                capacity_ton: spec.capacity_ton,
                avg_wait_s: Math.round(wait),
                avg_passage_s: Math.round(pass),
                water_per_ton_m3: Math.round(wpt * 100) / 100,
                throughput_tpd: Math.round(thr),
                conflict_rate: Math.round(confl * 1000) / 1000,
                efficiency_score: effScore,
                base_priority: spec.base_priority,
                color_hex: spec.color_hex,
                historical_usage: spec.historical_usage,
                dimensions: spec.length_max + '×' + spec.width_max + '×' + spec.draft_max + 'm'
            });
        });
        return out;
    }

    function renderCenter() {
        var center = document.getElementById('st-center');
        center.innerHTML =
            '<div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">' +
            '<div style="background:rgba(15,30,53,0.9);border:1px solid #1e3a5f;border-radius:8px;padding:14px;">' +
            '<h3 style="color:#4fc3f7;font-size:13px;margin-bottom:10px;">⏱ 等待时间 vs 通行时间（秒）</h3>' +
            '<canvas id="st-chart-time" style="width:100%;height:220px;"></canvas></div>' +
            '<div style="background:rgba(15,30,53,0.9);border:1px solid #1e3a5f;border-radius:8px;padding:14px;">' +
            '<h3 style="color:#4fc3f7;font-size:13px;margin-bottom:10px;">💧 吨水耗对比（m³/吨）</h3>' +
            '<canvas id="st-chart-water" style="width:100%;height:220px;"></canvas></div>' +
            '<div style="background:rgba(15,30,53,0.9);border:1px solid #1e3a5f;border-radius:8px;padding:14px;">' +
            '<h3 style="color:#4fc3f7;font-size:13px;margin-bottom:10px;">📊 效率分排名（综合指数）</h3>' +
            '<canvas id="st-chart-rank" style="width:100%;height:220px;"></canvas></div>' +
            '<div style="background:rgba(15,30,53,0.9);border:1px solid #1e3a5f;border-radius:8px;padding:14px;">' +
            '<h3 style="color:#4fc3f7;font-size:13px;margin-bottom:10px;">📦 日吞吐量曲线（吨/日）</h3>' +
            '<canvas id="st-chart-throughput" style="width:100%;height:220px;"></canvas></div>' +
            '</div>';
        setTimeout(function () {
            drawTimeChart(); drawWaterChart(); drawRankChart(); drawThroughputChart();
        }, 50);
    }

    function drawTimeChart() {
        var c = document.getElementById('st-chart-time'); if (!c) return;
        var ctx = c.getContext('2d');
        var w = c.width = c.clientWidth, h = c.height = c.clientHeight;
        var padL = 50, padR = 20, padT = 20, padB = 40;
        var gw = w - padL - padR, gh = h - padT - padB;
        var maxV = 0;
        reports.forEach(function (r) { maxV = Math.max(maxV, r.avg_wait_s, r.avg_passage_s); });
        maxV = Math.ceil(maxV / 300) * 300;
        ctx.strokeStyle = '#1e3a5f'; ctx.lineWidth = 1;
        for (var i = 0; i <= 4; i++) {
            var y = padT + gh * i / 4;
            ctx.beginPath(); ctx.moveTo(padL, y); ctx.lineTo(w - padR, y); ctx.stroke();
            ctx.fillStyle = '#90caf9'; ctx.font = '10px sans-serif'; ctx.textAlign = 'right';
            ctx.fillText(Math.round(maxV * (1 - i / 4)), padL - 6, y + 3);
        }
        var n = reports.length, bw = gw / n * 0.35, gap = gw / n * 0.3;
        reports.forEach(function (r, i) {
            var x = padL + gap / 2 + i * (bw * 2 + gap);
            var h1 = gh * r.avg_wait_s / maxV;
            var h2 = gh * r.avg_passage_s / maxV;
            ctx.fillStyle = '#ff7043'; ctx.fillRect(x, padT + gh - h1, bw, h1);
            ctx.fillStyle = '#4fc3f7'; ctx.fillRect(x + bw, padT + gh - h2, bw, h2);
            ctx.fillStyle = '#cfd8dc'; ctx.font = '10px sans-serif'; ctx.textAlign = 'center';
            var lbl = r.type_name.length > 3 ? r.type_name.substr(0, 3) : r.type_name;
            ctx.fillText(lbl, x + bw, h - padB + 12);
        });
        ctx.fillStyle = '#ff7043'; ctx.fillRect(padL + 4, 4, 10, 8);
        ctx.fillStyle = '#e3f2fd'; ctx.font = '10px sans-serif'; ctx.textAlign = 'left';
        ctx.fillText('等待', padL + 18, 12);
        ctx.fillStyle = '#4fc3f7'; ctx.fillRect(padL + 60, 4, 10, 8);
        ctx.fillStyle = '#e3f2fd'; ctx.fillText('通行', padL + 74, 12);
    }

    function drawWaterChart() {
        var c = document.getElementById('st-chart-water'); if (!c) return;
        var ctx = c.getContext('2d');
        var w = c.width = c.clientWidth, h = c.height = c.clientHeight;
        var padL = 90, padR = 30, padT = 16, padB = 20;
        var gw = w - padL - padR, gh = h - padT - padB;
        var maxV = 0;
        reports.forEach(function (r) { maxV = Math.max(maxV, r.water_per_ton_m3); });
        maxV = Math.ceil(maxV * 10) / 10 + 0.2;
        var n = reports.length, rh = gh / n * 0.65, rgap = gh / n * 0.35;
        reports.slice().sort(function (a, b) { return b.water_per_ton_m3 - a.water_per_ton_m3; }).forEach(function (r, i) {
            var y = padT + rgap / 2 + i * (rh + rgap);
            var bw = gw * r.water_per_ton_m3 / maxV;
            ctx.fillStyle = r.color_hex || '#4fc3f7';
            ctx.fillRect(padL, y, bw, rh);
            ctx.fillStyle = '#e3f2fd'; ctx.font = '10px sans-serif'; ctx.textAlign = 'right';
            ctx.fillText(r.type_name, padL - 8, y + rh / 2 + 3);
            ctx.textAlign = 'left';
            ctx.fillText(r.water_per_ton_m3.toFixed(2) + ' m³/t', padL + bw + 4, y + rh / 2 + 3);
        });
    }

    function drawRankChart() {
        var c = document.getElementById('st-chart-rank'); if (!c) return;
        var ctx = c.getContext('2d');
        var w = c.width = c.clientWidth, h = c.height = c.clientHeight;
        var padL = 50, padR = 30, padT = 16, padB = 40;
        var gw = w - padL - padR, gh = h - padT - padB;
        var sorted = reports.slice().sort(function (a, b) { return b.efficiency_score - a.efficiency_score; });
        var n = sorted.length, bw = gw / n * 0.7, gap = gw / n * 0.3;
        for (var i = 0; i <= 4; i++) {
            var y = padT + gh * i / 4;
            ctx.strokeStyle = '#1e3a5f'; ctx.lineWidth = 1;
            ctx.beginPath(); ctx.moveTo(padL, y); ctx.lineTo(w - padR, y); ctx.stroke();
            ctx.fillStyle = '#90caf9'; ctx.font = '10px sans-serif'; ctx.textAlign = 'right';
            ctx.fillText(Math.round(100 * (1 - i / 4)), padL - 6, y + 3);
        }
        sorted.forEach(function (r, i) {
            var x = padL + gap / 2 + i * (bw + gap);
            var bh = gh * r.efficiency_score / 100;
            var grad = ctx.createLinearGradient(0, padT + gh - bh, 0, padT + gh);
            grad.addColorStop(0, r.color_hex || '#4fc3f7');
            grad.addColorStop(1, '#0f2850');
            ctx.fillStyle = grad; ctx.fillRect(x, padT + gh - bh, bw, bh);
            ctx.strokeStyle = r.color_hex || '#4fc3f7'; ctx.lineWidth = 1.5;
            ctx.strokeRect(x, padT + gh - bh, bw, bh);
            ctx.fillStyle = '#ffd54f'; ctx.font = 'bold 11px sans-serif'; ctx.textAlign = 'center';
            ctx.fillText(r.efficiency_score, x + bw / 2, padT + gh - bh - 5);
            ctx.fillStyle = '#cfd8dc'; ctx.font = '10px sans-serif';
            var lbl = r.type_name.length > 3 ? r.type_name.substr(0, 3) : r.type_name;
            ctx.fillText(lbl, x + bw / 2, h - padB + 12);
            if (i === 0) { ctx.fillStyle = '#ffd700'; ctx.font = 'bold 14px sans-serif'; ctx.fillText('🏆', x + bw / 2, padT + gh - bh - 22); }
        });
    }

    function drawThroughputChart() {
        var c = document.getElementById('st-chart-throughput'); if (!c) return;
        var ctx = c.getContext('2d');
        var w = c.width = c.clientWidth, h = c.height = c.clientHeight;
        var padL = 60, padR = 20, padT = 30, padB = 40;
        var gw = w - padL - padR, gh = h - padT - padB;
        var maxV = 0;
        reports.forEach(function (r) { maxV = Math.max(maxV, r.throughput_tpd); });
        maxV = Math.ceil(maxV / 500) * 500;
        for (var i = 0; i <= 4; i++) {
            var y = padT + gh * i / 4;
            ctx.strokeStyle = '#1e3a5f'; ctx.beginPath(); ctx.moveTo(padL, y); ctx.lineTo(w - padR, y); ctx.stroke();
            ctx.fillStyle = '#90caf9'; ctx.font = '10px sans-serif'; ctx.textAlign = 'right';
            ctx.fillText(Math.round(maxV * (1 - i / 4)), padL - 6, y + 3);
        }
        var n = reports.length;
        reports.forEach(function (r, i) {
            var x = padL + (gw / (n - 1 || 1)) * i;
            var y = padT + gh * (1 - r.throughput_tpd / maxV);
            r._x = x; r._y = y;
        });
        if (n > 1) {
            ctx.strokeStyle = '#4fc3f7'; ctx.lineWidth = 2; ctx.beginPath();
            ctx.moveTo(reports[0]._x, reports[0]._y);
            for (var j = 1; j < n; j++) {
                var xc = (reports[j - 1]._x + reports[j]._x) / 2;
                ctx.quadraticCurveTo(xc, reports[j - 1]._y, xc, (reports[j - 1]._y + reports[j]._y) / 2);
                ctx.quadraticCurveTo(xc, reports[j]._y, reports[j]._x, reports[j]._y);
            }
            ctx.stroke();
        }
        reports.forEach(function (r) {
            ctx.fillStyle = 'rgba(79,195,247,0.15)';
            ctx.beginPath(); ctx.moveTo(r._x, padT + gh); ctx.lineTo(r._x, r._y);
            ctx.arc(r._x, r._y, 4, 0, Math.PI * 2); ctx.fill();
            ctx.fillStyle = r.color_hex || '#fff';
            ctx.beginPath(); ctx.arc(r._x, r._y, 5, 0, Math.PI * 2); ctx.fill();
            ctx.strokeStyle = '#fff'; ctx.lineWidth = 1.5; ctx.stroke();
            ctx.fillStyle = '#e3f2fd'; ctx.font = '10px sans-serif'; ctx.textAlign = 'center';
            ctx.fillText(r.throughput_tpd + 't', r._x, r._y - 10);
            var lbl = r.type_name.length > 3 ? r.type_name.substr(0, 3) : r.type_name;
            ctx.fillText(lbl, r._x, h - padB + 12);
        });
    }

    function renderRight() {
        var right = document.getElementById('st-right');
        var sorted = reports.slice().sort(function (a, b) { return b.efficiency_score - a.efficiency_score; });
        var top = sorted[0], worst = sorted[sorted.length - 1];
        right.innerHTML =
            '<div style="color:#4fc3f7;font-size:13px;font-weight:600;margin-bottom:10px;">🎯 综合分析</div>' +
            '<div style="background:rgba(30,58,95,0.3);border-radius:6px;padding:12px;margin-bottom:12px;">' +
            '<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px;">' +
            '<span style="color:#ffd54f;font-size:12px;">🏆 最优船型</span>' +
            '<span style="color:' + (top ? top.color_hex : '#fff') + ';font-weight:600;font-size:12px;">' + (top ? top.type_name : '-') + '</span></div>' +
            (top ? '<div style="color:#90caf9;font-size:11px;line-height:1.7;">效率分: ' + top.efficiency_score + '<br>日吞吐: ' + top.throughput_tpd + '吨<br>吨水耗: ' + top.water_per_ton_m3 + 'm³/t</div>' : '') +
            '</div>' +
            (worst && worst !== top ?
                '<div style="background:rgba(244,67,54,0.08);border:1px solid rgba(244,67,54,0.2);border-radius:6px;padding:12px;margin-bottom:12px;">' +
                '<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px;">' +
                '<span style="color:#ef5350;font-size:12px;">⚠ 待优化</span>' +
                '<span style="color:' + worst.color_hex + ';font-weight:600;font-size:12px;">' + worst.type_name + '</span></div>' +
                '<div style="color:#90caf9;font-size:11px;line-height:1.7;">效率分: ' + worst.efficiency_score + '<br>冲突率: ' + Math.round(worst.conflict_rate * 100) + '%<br>建议: 错峰调度+优先配大闸</div>' +
                '</div>' : '') +
            '<div style="color:#4fc3f7;font-size:13px;font-weight:600;margin-bottom:10px;">📐 七维雷达对比</div>' +
            '<canvas id="st-radar" style="width:100%;height:260px;margin-bottom:12px;"></canvas>' +
            '<div style="color:#4fc3f7;font-size:13px;font-weight:600;margin-bottom:10px;">📋 船型详情</div>' +
            '<div id="st-detail" style="font-size:11px;line-height:1.7;color:#90caf9;">悬停或点击左侧船型查看详情</div>';
        setTimeout(function () {
            drawRadar();
            bindDetail();
        }, 50);
    }

    function drawRadar() {
        var c = document.getElementById('st-radar'); if (!c) return;
        var ctx = c.getContext('2d');
        var w = c.width = c.clientWidth, h = c.height = c.clientHeight;
        var cx = w / 2, cy = h / 2, R = Math.min(w, h) / 2 - 30;
        var dims = ['效率', '吞吐', '节水', '低冲突', '快速', '大容量', '优先级'];
        var n = dims.length;
        ctx.strokeStyle = '#1e3a5f'; ctx.lineWidth = 1;
        for (var lvl = 1; lvl <= 4; lvl++) {
            var r = R * lvl / 4;
            ctx.beginPath();
            for (var k = 0; k < n; k++) {
                var ang = -Math.PI / 2 + 2 * Math.PI * k / n;
                var x = cx + r * Math.cos(ang), y = cy + r * Math.sin(ang);
                if (k === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y);
            }
            ctx.closePath(); ctx.stroke();
        }
        for (var k2 = 0; k2 < n; k2++) {
            var ang2 = -Math.PI / 2 + 2 * Math.PI * k2 / n;
            ctx.beginPath();
            ctx.moveTo(cx, cy); ctx.lineTo(cx + R * Math.cos(ang2), cy + R * Math.sin(ang2)); ctx.stroke();
            ctx.fillStyle = '#cfd8dc'; ctx.font = '11px sans-serif'; ctx.textAlign = 'center';
            var lx = cx + (R + 16) * Math.cos(ang2), ly = cy + (R + 16) * Math.sin(ang2);
            ctx.fillText(dims[k2], lx, ly + 4);
        }
        var maxW = 0, maxT = 0;
        reports.forEach(function (r) { maxW = Math.max(maxW, r.water_per_ton_m3); maxT = Math.max(maxT, r.throughput_tpd); });
        reports.forEach(function (r, idx) {
            var vals = [
                r.efficiency_score / 100,
                r.throughput_tpd / maxT,
                1 - Math.min(1, r.water_per_ton_m3 / maxW),
                1 - r.conflict_rate,
                1 - Math.min(1, r.avg_passage_s / 2400),
                Math.min(1, r.capacity_ton / 200),
                r.base_priority / 6
            ];
            ctx.strokeStyle = r.color_hex || '#4fc3f7'; ctx.lineWidth = 1.5;
            ctx.fillStyle = (r.color_hex || '#4fc3f7') + '33';
            ctx.beginPath();
            for (var m = 0; m < n; m++) {
                var ang3 = -Math.PI / 2 + 2 * Math.PI * m / n;
                var rr = R * vals[m];
                var px = cx + rr * Math.cos(ang3), py = cy + rr * Math.sin(ang3);
                if (m === 0) ctx.moveTo(px, py); else ctx.lineTo(px, py);
            }
            ctx.closePath(); ctx.fill(); ctx.stroke();
        });
        var lgx = 10, lgy = 8;
        reports.forEach(function (r) {
            ctx.fillStyle = r.color_hex || '#4fc3f7';
            ctx.fillRect(lgx, lgy, 10, 10);
            ctx.fillStyle = '#cfd8dc'; ctx.font = '10px sans-serif'; ctx.textAlign = 'left';
            ctx.fillText(r.type_name, lgx + 14, lgy + 9);
            lgy += 14;
        });
    }

    function bindDetail() {
        var list = document.querySelectorAll('#st-type-list [data-type]');
        list.forEach(function (row) {
            row.addEventListener('mouseenter', function () { showDetail(row.dataset.type); });
            row.addEventListener('click', function () { showDetail(row.dataset.type); });
        });
        if (reports[0]) showDetail(reports[0].ship_type);
    }

    function showDetail(type) {
        var r = reports.find(function (x) { return x.ship_type === type; });
        if (!r) return;
        var el = document.getElementById('st-detail');
        if (!el) return;
        el.innerHTML =
            '<div style="background:rgba(30,58,95,0.3);border-radius:6px;padding:12px;border-left:3px solid ' + (r.color_hex || '#4fc3f7') + ';">' +
            '<div style="color:' + (r.color_hex || '#fff') + ';font-weight:600;font-size:13px;margin-bottom:8px;">' + r.type_name + '</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#4fc3f7;">尺寸:</b> ' + r.dimensions + '</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#4fc3f7;">载重:</b> ' + r.capacity_ton + '吨</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#4fc3f7;">优先级:</b> P' + r.base_priority + '</div>' +
            '<div style="color:#90caf9;font-size:11px;margin:6px 0;padding-top:6px;border-top:1px solid #1e3a5f;"><b style="color:#ff7043;">等待:</b> ' + r.avg_wait_s + 's · <b style="color:#4fc3f7;">通行:</b> ' + r.avg_passage_s + 's</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#66bb6a;">吨水耗:</b> ' + r.water_per_ton_m3 + 'm³/t</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#ffd54f;">日吞吐:</b> ' + r.throughput_tpd + '吨</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#ef5350;">冲突率:</b> ' + Math.round(r.conflict_rate * 100) + '%</div>' +
            '<div style="color:#90caf9;font-size:11px;"><b style="color:#ab47bc;">综合分:</b> <span style="color:#ffd54f;font-weight:600;font-size:14px;">' + r.efficiency_score + '</span></div>' +
            '<div style="color:#78909c;font-size:10px;margin-top:8px;padding-top:6px;border-top:1px solid #1e3a5f;font-style:italic;">📜 ' + r.historical_usage + '</div>' +
            '</div>';
    }

    function show() {
        var v = document.getElementById('ship-view'); if (v) v.style.display = 'flex';
    }
    function hide() {
        var v = document.getElementById('ship-view'); if (v) v.style.display = 'none';
    }
    function setGate(gid) {
        gateId = gid;
        var sel = document.getElementById('st-gate'); if (sel) sel.value = gid;
    }

    return { init: init, show: show, hide: hide, setGate: setGate };
})();
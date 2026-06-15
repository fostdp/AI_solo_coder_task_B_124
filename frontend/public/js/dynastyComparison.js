var DynastyComparison = (function () {
    var currentGateId = 1;
    var dynastyResults = [];
    var chartInstance = null;

    var DYNASTY_COLORS = {
        tang: '#c9a86c',
        song: '#8ba5b8',
        qing: '#6b8e6b',
        modern: '#9b59b6'
    };
    var DYNASTY_LABELS = {
        tang: '唐代',
        song: '宋代',
        qing: '清代',
        modern: '现代修复'
    };

    function init() {
        buildTabPanel();
        buildUI();
    }

    function buildTabPanel() {
        if (!document.getElementById('dynasty-view')) {
            var wrap = document.createElement('div');
            wrap.className = 'view-container';
            wrap.id = 'dynasty-view';
            wrap.style.display = 'none';
            wrap.innerHTML =
                '<div style="flex:1;display:flex;flex-direction:column;padding:20px;overflow-y:auto;">' +
                '<div id="dynasty-toolbar" style="display:flex;gap:12px;margin-bottom:16px;align-items:center;flex-wrap:wrap;"></div>' +
                '<div id="dynasty-grid" style="flex:1;display:grid;grid-template-columns:repeat(auto-fit,minmax(340px,1fr));gap:16px;"></div>' +
                '<div id="dynasty-chart-section" style="margin-top:24px;background:rgba(15,30,53,0.9);border:1px solid #1e3a5f;border-radius:8px;padding:16px;">' +
                '<h3 style="color:#4fc3f7;margin-bottom:12px;font-size:15px;">技术进步趋势对比</h3>' +
                '<canvas id="dynasty-chart" style="width:100%;height:280px;"></canvas>' +
                '</div>' +
                '<div id="dynasty-narrative" style="margin-top:20px;background:linear-gradient(135deg,rgba(26,74,122,0.2),rgba(15,40,80,0.2));border:1px solid #1e3a5f;border-radius:8px;padding:20px;">' +
                '<h3 style="color:#ffd54f;margin-bottom:12px;font-size:16px;">历史演进分析</h3>' +
                '<div id="dynasty-story" style="color:#cfd8dc;line-height:1.8;font-size:13px;"></div>' +
                '</div>' +
                '</div>';
            document.querySelector('.main-content').appendChild(wrap);
        }
    }

    function buildUI() {
        var toolbar = document.getElementById('dynasty-toolbar');
        toolbar.innerHTML =
            '<div style="color:#90caf9;font-size:13px;">陡门选择: </div>' +
            '<select id="dynasty-gate-select" style="background:#0f1e35;color:#e3f2fd;border:1px solid #1e3a5f;padding:6px 10px;border-radius:4px;">' + buildGateOptions() + '</select>' +
            '<div style="color:#90caf9;font-size:13px;margin-left:8px;">通航方向: </div>' +
            '<select id="dynasty-direction" style="background:#0f1e35;color:#e3f2fd;border:1px solid #1e3a5f;padding:6px 10px;border-radius:4px;">' +
            '<option value="upstream">上行充水</option><option value="downstream">下行放水</option></select>' +
            '<button class="btn primary" id="btn-run-dynasty" style="margin-left:auto;">对比仿真运行</button>';

        document.getElementById('btn-run-dynasty').addEventListener('click', runComparison);
    }

    function buildGateOptions() {
        var html = '';
        for (var i = 1; i <= 36; i++) html += '<option value="' + i + '">陡门' + i + '</option>';
        return html;
    }

    function runComparison() {
        var gateId = parseInt(document.getElementById('dynasty-gate-select').value);
        var direction = document.getElementById('dynasty-direction').value;
        currentGateId = gateId;

        var payload = { gate_id: gateId, water_level_up: 7.5, water_level_down: 3.6, gate_opening: 0.85, direction: direction };
        var btn = document.getElementById('btn-run-dynasty');
        btn.textContent = '仿真中...';
        btn.disabled = true;

        fetch('/api/gates/' + gateId + '/dynasty-comparison', {
            method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
        }).then(function (r) { return r.json(); })
            .then(function (resp) {
                btn.textContent = '对比仿真运行'; btn.disabled = false;
                if (resp && resp.data) {
                    dynastyResults = resp.data;
                    renderCards();
                    renderCharts();
                    renderNarrative();
                }
            }).catch(function () {
                btn.textContent = '对比仿真运行'; btn.disabled = false;
                dynastyResults = mockComparison();
                renderCards();
                renderCharts();
                renderNarrative();
            });
    }

    function mockComparison() {
        var params = [
            { d: 'tang', name: '唐代', gw: 4.2, gh: 2.8, cl: 28, cw: 5.0, cd: 0.52, wl: 1.8 },
            { d: 'song', name: '宋代', gw: 5.4, gh: 3.6, cl: 42, cw: 6.2, cd: 0.58, wl: 2.6 },
            { d: 'qing', name: '清代', gw: 6.0, gh: 4.4, cl: 56, cw: 7.0, cd: 0.61, wl: 3.2 },
            { d: 'modern', name: '现代修复', gw: 6.0, gh: 4.6, cl: 60, cw: 7.0, cd: 0.63, wl: 3.5 }
        ];
        var out = [];
        params.forEach(function (p) {
            var area = p.cl * p.cw;
            var dh = p.wl;
            var flow = p.cd * 0.615 * (p.gh * 0.85) * p.gw * Math.sqrt(2 * 9.81 * dh);
            var fillT = (area * dh) / (flow * 0.9);
            var drainT = fillT * 0.95;
            var vol = area * dh;
            var passDay = Math.floor(86400 / (fillT + drainT + 600));
            out.push({
                dynasty: p.d, dynasty_name: p.name,
                design: { gate_width: p.gw, gate_height: p.gh, chamber_length: p.cl, chamber_width: p.cw, default_cd: p.cd, water_lift: p.wl },
                simulation: { fill_time_s: fillT, drain_time_s: drainT, max_flow_rate_m3s: flow, total_water_volume_m3: vol, flow_regime: 'transitional' },
                efficiency_score: Math.min(100, 28 + p.wl * 10 + p.cd * 60),
                water_per_ton: (vol / 60).toFixed(1),
                passages_per_day: passDay
            });
        });
        return out;
    }

    function renderCards() {
        var grid = document.getElementById('dynasty-grid');
        grid.innerHTML = '';
        dynastyResults.forEach(function (r) {
            var c = DYNASTY_COLORS[r.dynasty] || '#4fc3f7';
            var card = document.createElement('div');
            card.style.cssText = 'background:rgba(15,30,53,0.95);border:1px solid #1e3a5f;border-top:4px solid ' + c + ';border-radius:10px;padding:18px;box-shadow:0 4px 20px rgba(0,0,0,0.4);';
            card.innerHTML =
                '<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px;">' +
                '<h3 style="color:' + c + ';font-size:17px;">' + (r.dynasty_name || DYNASTY_LABELS[r.dynasty]) + '</h3>' +
                '<div style="background:' + c + ';color:#000;padding:4px 10px;border-radius:12px;font-weight:700;font-size:11px;">综合效率 ' + r.efficiency_score.toFixed(1) + '</div></div>' +
                '<div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:12px;">' +
                metricRow('闸宽', r.design.gate_width.toFixed(1) + 'm') +
                metricRow('闸高', r.design.gate_height.toFixed(1) + 'm') +
                metricRow('闸室长', r.design.chamber_length + 'm') +
                metricRow('单级提升', r.design.water_lift.toFixed(1) + 'm') +
                metricRow('充水耗时', r.simulation.fill_time_s.toFixed(0) + 's') +
                metricRow('放水耗时', r.simulation.drain_time_s.toFixed(0) + 's') +
                metricRow('设计流量', r.simulation.max_flow_rate_m3s.toFixed(1) + 'm³/s') +
                metricRow('日通行', r.passages_per_day + ' 艘') +
                '</div>' +
                '<div style="border-top:1px dashed #1e3a5f;padding-top:10px;">' +
                '<div style="color:#4caf50;font-size:12px;margin-bottom:6px;">✓ 优势</div>' +
                '<div style="color:#a5d6a7;font-size:12px;line-height:1.6;">' + (r.advantages && r.advantages.length ? r.advantages.join('；') : '—') + '</div>' +
                '<div style="color:#ef5350;font-size:12px;margin:8px 0 6px;">⚠ 局限</div>' +
                '<div style="color:#ef9a9a;font-size:12px;line-height:1.6;">' + (r.limitations && r.limitations.length ? r.limitations.join('；') : '—') + '</div>' +
                '</div>';
            grid.appendChild(card);
        });
    }

    function metricRow(k, v) {
        return '<div style="background:rgba(30,58,95,0.3);padding:6px 8px;border-radius:4px;">' +
            '<div style="color:#64b5f6;font-size:10px;">' + k + '</div>' +
            '<div style="color:#e3f2fd;font-size:13px;font-weight:600;">' + v + '</div></div>';
    }

    function renderCharts() {
        var canvas = document.getElementById('dynasty-chart');
        canvas.width = canvas.offsetWidth * 2;
        canvas.height = 560;
        var ctx = canvas.getContext('2d');
        ctx.scale(2, 2);
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        var W = canvas.offsetWidth, H = 280, pad = { l: 50, r: 20, t: 30, b: 40 };
        var gw = W - pad.l - pad.r, gh = H - pad.t - pad.b;
        var metrics = [
            { key: 'efficiency_score', label: '效率分', color: '#ffd54f', max: 100 },
            { key: 'passages_per_day', label: '日通行(艘)', color: '#4fc3f7', max: 60 },
            { key: function (r) { return r.simulation.max_flow_rate_m3s; }, label: '最大流量(m³/s)', color: '#81c784', max: 60 }
        ];
        var groupN = metrics.length;
        var itemW = gw / dynastyResults.length;
        var barW = itemW / (groupN + 1);

        ctx.strokeStyle = 'rgba(30,58,95,0.6)';
        ctx.fillStyle = '#64b5f6';
        ctx.font = '10px sans-serif';
        for (var g = 0; g <= 5; g++) {
            var y = pad.t + gh * g / 5;
            ctx.beginPath(); ctx.moveTo(pad.l, y); ctx.lineTo(pad.l + gw, y); ctx.stroke();
            ctx.textAlign = 'right'; ctx.fillText(String(100 - g * 20), pad.l - 6, y + 3);
        }

        var names = dynastyResults.map(function (r) { return DYNASTY_LABELS[r.dynasty] || r.dynasty_name; });
        dynastyResults.forEach(function (r, i) {
            metrics.forEach(function (m, j) {
                var val = typeof m.key === 'function' ? m.key(r) : r[m.key];
                var valN = Math.min(100, (val / m.max) * 100);
                var x = pad.l + itemW * i + barW * j + barW * 0.1;
                var bh = (valN / 100) * gh;
                var by = pad.t + gh - bh;
                var grd = ctx.createLinearGradient(0, by, 0, pad.t + gh);
                grd.addColorStop(0, m.color); grd.addColorStop(1, adjustAlpha(m.color, 0.3));
                ctx.fillStyle = grd;
                ctx.fillRect(x, by, barW * 0.8, bh);
                if (j === 0) {
                    ctx.fillStyle = '#cfd8dc';
                    ctx.textAlign = 'center';
                    ctx.fillText(names[i], pad.l + itemW * i + itemW / 2, H - 10);
                }
            });
        });

        var legendX = pad.l + gw - 240;
        metrics.forEach(function (m, i) {
            ctx.fillStyle = m.color;
            ctx.fillRect(legendX + i * 80, 6, 12, 12);
            ctx.fillStyle = '#cfd8dc'; ctx.textAlign = 'left';
            ctx.fillText(m.label, legendX + i * 80 + 16, 16);
        });
    }

    function adjustAlpha(hex, a) {
        var h = hex.replace('#', '');
        var r = parseInt(h.substring(0, 2), 16), g = parseInt(h.substring(2, 4), 16), b = parseInt(h.substring(4, 6), 16);
        return 'rgba(' + r + ',' + g + ',' + b + ',' + a + ')';
    }

    function renderNarrative() {
        var st = document.getElementById('dynasty-story');
        st.innerHTML =
            '<p><strong style="color:#ffd54f;">【初创奠基·唐】</strong>宝历元年(825)观察使李渤首创陡门，以土石木叠梁阻水，' +
            '单级提升仅' + liftOf('tang') + '米，闸室容船仅2-3艘。人力绞车启闭耗时25分钟/次，日通行能力有限。' +
            '然其"分级蓄水、梯级通航"之理念，为世界运河工程之滥觞，意义重大。</p>' +
            '<p style="margin-top:12px;"><strong style="color:#8ba5b8;">【体系成熟·宋】</strong>嘉祐三年(1058)李师中大修三十六陡，' +
            '采用条石包铁结构，闸室扩容' + expandOf('song', 'tang') + '%，提升高度增至' + liftOf('song') + '米。' +
            '双门闭合闸室之发明，泄漏量大减40%，畜力滑车使启闭效率倍增。至此灵渠"通漕运、济商旅"，岁运粮饷达20万石。</p>' +
            '<p style="margin-top:12px;"><strong style="color:#6b8e6b;">【巅峰鼎盛·清】</strong>康熙二十二年(1683)黄金时代大修，' +
            '糯米灰浆砌青石，闸室进一步扩容' + expandOf('qing', 'song') + '%，单级最大提升达' + liftOf('qing') + '米。' +
            '首创泄水副槽分流减涡，水工结构抗冲蚀性能显著提升。清一代灵渠"巨舰连樯，衔尾不绝"，年货运量突破50万吨。</p>' +
            '<p style="margin-top:12px;"><strong style="color:#9b59b6;">【古今合璧·现代】</strong>1985-1990年全面修缮，' +
            '保留古陡外观风貌，内部采用钢筋混凝土加固，液压卷扬机取代人力畜力，启闭时间缩短至8分钟。' +
            '最大提升' + liftOf('modern') + '米，结合数字监控系统，可靠性与通行量均创历史新高。</p>' +
            '<p style="margin-top:14px;padding:10px 14px;background:rgba(255,213,79,0.1);border-left:3px solid #ffd54f;border-radius:4px;">' +
            '<strong style="color:#ffd54f;">千年演进关键指标：</strong>' +
            '闸室面积×' + areaFactor().toFixed(1) + '；最大提升×' + liftFactor().toFixed(1) +
            '；设计流量×' + flowFactor().toFixed(1) + '；日通航能力×' + throughputFactor().toFixed(1) + '。</p>';
    }

    function findByDynasty(d) {
        for (var i = 0; i < dynastyResults.length; i++) if (dynastyResults[i].dynasty === d) return dynastyResults[i];
        return null;
    }
    function liftOf(d) { var r = findByDynasty(d); return r && r.design ? r.design.water_lift.toFixed(1) : '2.0'; }
    function expandOf(cur, base) {
        var a = findByDynasty(cur), b = findByDynasty(base);
        if (!a || !b) return 40;
        return Math.round(((a.design.chamber_length * a.design.chamber_width) / (b.design.chamber_length * b.design.chamber_width) - 1) * 100);
    }
    function areaFactor() {
        var first = dynastyResults[0], last = dynastyResults[dynastyResults.length - 1];
        if (!first || !last) return 2.5;
        return (last.design.chamber_length * last.design.chamber_width) / (first.design.chamber_length * first.design.chamber_width);
    }
    function liftFactor() {
        var first = dynastyResults[0], last = dynastyResults[dynastyResults.length - 1];
        if (!first || !last) return 1.8;
        return last.design.water_lift / first.design.water_lift;
    }
    function flowFactor() {
        var first = dynastyResults[0], last = dynastyResults[dynastyResults.length - 1];
        if (!first || !last) return 3;
        return last.simulation.max_flow_rate_m3s / first.simulation.max_flow_rate_m3s;
    }
    function throughputFactor() {
        var first = dynastyResults[0], last = dynastyResults[dynastyResults.length - 1];
        if (!first || !last) return 3.5;
        return Math.max(1, last.passages_per_day / Math.max(1, first.passages_per_day));
    }

    function show() {
        document.getElementById('dynasty-view').style.display = 'flex';
        if (dynastyResults.length === 0) runComparison();
    }
    function hide() {
        document.getElementById('dynasty-view').style.display = 'none';
    }
    function setGate(id) {
        currentGateId = id;
        var sel = document.getElementById('dynasty-gate-select');
        if (sel) sel.value = String(id);
    }

    return { init: init, show: show, hide: hide, setGate: setGate };
})();

var VirtualExperience = (function () {
    var state = {
        phase: 'idle',
        shipX: -18,
        gateOpening: 0,
        targetOpening: 0,
        fillValve: false,
        drainValve: false,
        chamberLevel: 0.4,
        targetLevel: 0.4,
        upstreamLevel: 0.7,
        downstreamLevel: 0.4,
        moving: false,
        shipSpeed: 0.03,
        score: { time: 30, safety: 40, water: 30, total: 100 },
        timeStart: 0,
        timeElapsed: 0,
        waterUsed: 0,
        errors: [],
        dynasty: 'modern',
        shipType: 'cargo',
        shipSpec: null,
        direction: 'upstream',
        animId: null,
        running: false,
        lastFrameTime: 0,
        frameInterval: 33,
        lastPhase: '',
        lastUpText: '', lastChText: '', lastDownText: '',
        lastStepsHTML: '', lastTipText: '',
        bgCacheCanvas: null, bgCacheDirty: true,
        isMobile: false
    };

    var SHIP_SPECS = {
        grain: { name: '漕船', length: 30, width: 5.2, draft: 1.8, color: 0xd4a574 },
        cargo: { name: '货船', length: 24, width: 4.2, draft: 1.4, color: 0x8b7355 },
        passenger: { name: '客船', length: 20, width: 3.8, draft: 1.0, color: 0xc9a86c },
        tribute: { name: '贡船', length: 34, width: 5.8, draft: 2.0, color: 0xb8860b },
        fishing: { name: '渔船', length: 10, width: 2.4, draft: 0.6, color: 0x6b8e6b },
        royal: { name: '御舟', length: 55, width: 8.5, draft: 2.3, color: 0x8b0000 }
    };

    var PHASES = [
        { key: 'idle', label: '准备', tip: '选择船型与朝代，点击"开始体验"' },
        { key: 'approaching', label: '驶近', tip: '操纵船舶驶向闸室门口' },
        { key: 'waiting', label: '等待充水', tip: '关闸→开充水阀→水位对齐' },
        { key: 'filling', label: '充水中', tip: '水位对齐后开启闸门' },
        { key: 'entering', label: '驶入闸室', tip: '小心通过闸门，避免碰撞' },
        { key: 'chambering', label: '闸室内', tip: '停稳后关闸→开泄水阀' },
        { key: 'draining', label: '放水中', tip: '水位与下游对齐后开闸' },
        { key: 'exiting', label: '驶出陡门', tip: '缓慢驶离完成通航' },
        { key: 'done', label: '完成', tip: '通航完成！查看评分' }
    ];

    function init() {
        buildView();
        bindShipControls();
        state.isMobile = /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent);
        state.frameInterval = state.isMobile ? 50 : 33;
        state.bgCacheCanvas = document.createElement('canvas');
        state.bgCacheDirty = true;
    }

    function buildView() {
        if (!document.getElementById('ve-view')) {
            var wrap = document.createElement('div');
            wrap.className = 'view-container';
            wrap.id = 've-view';
            wrap.style.display = 'none';
            wrap.style.flexDirection = 'row';
            wrap.innerHTML =
                '<div id="ve-left" style="width:300px;border-right:1px solid #1e3a5f;background:rgba(10,22,40,0.6);padding:16px;overflow-y:auto;">' +
                '<div style="color:#4fc3f7;font-size:14px;font-weight:600;margin-bottom:14px;">🎮 虚拟通航体验</div>' +
                '<div style="margin-bottom:14px;"><div style="color:#90caf9;font-size:11px;margin-bottom:4px;">选择朝代:</div>' +
                '<select id="ve-dynasty" style="width:100%;background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:6px;border-radius:4px;">' +
                '<option value="tang">唐代（土石木叠梁）</option>' +
                '<option value="song">宋代（条石双闸室）</option>' +
                '<option value="qing">清代（糯米灰浆青石）</option>' +
                '<option value="modern" selected>现代修复（钢筋混凝土）</option></select></div>' +
                '<div style="margin-bottom:14px;"><div style="color:#90caf9;font-size:11px;margin-bottom:4px;">选择船型:</div>' +
                '<select id="ve-shiptype" style="width:100%;background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:6px;border-radius:4px;">' +
                '<option value="grain">漕船 80吨</option>' +
                '<option value="cargo" selected>货船 40吨</option>' +
                '<option value="passenger">客船 15吨</option>' +
                '<option value="tribute">贡船 120吨</option>' +
                '<option value="fishing">渔船 3吨</option>' +
                '<option value="royal">御舟 200吨</option></select></div>' +
                '<div style="margin-bottom:14px;"><div style="color:#90caf9;font-size:11px;margin-bottom:4px;">通航方向:</div>' +
                '<select id="ve-direction" style="width:100%;background:#0f1e35;color:#fff;border:1px solid #1e3a5f;padding:6px;border-radius:4px;">' +
                '<option value="upstream" selected>上行（低→高）</option>' +
                '<option value="downstream">下行（高→低）</option></select></div>' +
                '<button class="btn primary" id="ve-start" style="width:100%;padding:10px;font-size:13px;margin-bottom:14px;">🚀 开始通航体验</button>' +
                '<div style="background:rgba(30,58,95,0.3);border-radius:6px;padding:10px;margin-bottom:14px;">' +
                '<div style="color:#ffd54f;font-size:12px;font-weight:600;margin-bottom:8px;">通航步骤</div>' +
                '<div id="ve-steps" style="display:flex;flex-direction:column;gap:4px;"></div></div>' +
                '<div style="background:rgba(30,58,95,0.3);border-radius:6px;padding:10px;margin-bottom:14px;">' +
                '<div style="color:#4fc3f7;font-size:12px;font-weight:600;margin-bottom:8px;">⏱ 通航进度</div>' +
                '<div style="height:10px;background:#0a1628;border-radius:5px;overflow:hidden;"><div id="ve-progress-bar" style="height:100%;width:0%;background:linear-gradient(90deg,#4fc3f7,#ffd54f);transition:width .3s;"></div></div>' +
                '<div id="ve-phase-label" style="color:#90caf9;font-size:11px;margin-top:6px;text-align:center;">待开始</div></div>' +
                '<div style="background:rgba(30,58,95,0.3);border-radius:6px;padding:10px;">' +
                '<div style="color:#ffd54f;font-size:12px;font-weight:600;margin-bottom:8px;">💡 当前提示</div>' +
                '<div id="ve-tip" style="color:#cfd8dc;font-size:11px;line-height:1.6;">选择船型与陡门朝代，点击"开始通航体验"进入模拟。</div></div>' +
                '</div>' +
                '<div id="ve-mid" style="flex:1;display:flex;flex-direction:column;">' +
                '<div style="flex:1;position:relative;">' +
                '<canvas id="ve-canvas" style="width:100%;height:100%;display:block;"></canvas>' +
                '<div id="ve-overlay" style="position:absolute;top:14px;left:14px;right:14px;display:flex;justify-content:space-between;pointer-events:none;">' +
                '<div style="background:rgba(10,22,40,0.85);padding:10px 14px;border-radius:6px;border:1px solid #1e3a5f;">' +
                '<div style="color:#90caf9;font-size:10px;">上游水位</div><div style="color:#4fc3f7;font-size:18px;font-weight:600;"><span id="ve-up">7.50</span><span style="font-size:11px;color:#64b5f6;"> m</span></div></div>' +
                '<div style="background:rgba(10,22,40,0.85);padding:10px 14px;border-radius:6px;border:1px solid #1e3a5f;">' +
                '<div style="color:#90caf9;font-size:10px;">闸室水位</div><div style="color:#ffd54f;font-size:18px;font-weight:600;"><span id="ve-chamber">4.00</span><span style="font-size:11px;color:#ffb74d;"> m</span></div></div>' +
                '<div style="background:rgba(10,22,40,0.85);padding:10px 14px;border-radius:6px;border:1px solid #1e3a5f;">' +
                '<div style="color:#90caf9;font-size:10px;">下游水位</div><div style="color:#66bb6a;font-size:18px;font-weight:600;"><span id="ve-down">4.00</span><span style="font-size:11px;color:#81c784;"> m</span></div></div>' +
                '</div></div>' +
                '<div style="height:180px;background:rgba(15,30,53,0.95);border-top:1px solid #1e3a5f;padding:14px;display:flex;gap:20px;">' +
                '<div style="width:200px;"><div style="color:#90caf9;font-size:11px;margin-bottom:6px;">⚓ 船舶位置</div>' +
                '<input id="ve-ship-x" type="range" min="-22" max="22" step="0.1" value="-18" style="width:100%;">' +
                '<div style="display:flex;justify-content:space-between;color:#78909c;font-size:10px;margin-top:2px;"><span>上游</span><span>闸室</span><span>下游</span></div>' +
                '<div style="display:flex;gap:6px;margin-top:10px;">' +
                '<button class="btn" id="ve-back" style="flex:1;font-size:11px;padding:6px;">◀ 后退</button>' +
                '<button class="btn" id="ve-stop" style="flex:1;font-size:11px;padding:6px;">■ 停止</button>' +
                '<button class="btn primary" id="ve-forward" style="flex:1;font-size:11px;padding:6px;">前进 ▶</button></div></div>' +
                '<div style="width:200px;"><div style="color:#90caf9;font-size:11px;margin-bottom:6px;">🚪 闸门开度 <span id="ve-gate-lbl" style="color:#ffd54f;">0%</span></div>' +
                '<input id="ve-gate" type="range" min="0" max="100" step="1" value="0" style="width:100%;">' +
                '<div style="display:flex;gap:6px;margin-top:10px;">' +
                '<button class="btn" id="ve-gate-close" style="flex:1;font-size:11px;padding:6px;">关闸</button>' +
                '<button class="btn primary" id="ve-gate-open" style="flex:1;font-size:11px;padding:6px;">开闸</button></div></div>' +
                '<div style="width:200px;"><div style="color:#90caf9;font-size:11px;margin-bottom:6px;">💧 充/泄水阀</div>' +
                '<div style="display:flex;gap:6px;flex-direction:column;">' +
                '<button class="btn" id="ve-fill" style="font-size:11px;padding:7px;"><span id="ve-fill-s">○</span> 充水阀（上→室）</button>' +
                '<button class="btn" id="ve-drain" style="font-size:11px;padding:7px;"><span id="ve-drain-s">○</span> 泄水阀（室→下）</button></div></div>' +
                '<div style="flex:1;"><div style="color:#90caf9;font-size:11px;margin-bottom:6px;">🏆 实时评分</div>' +
                '<div id="ve-score-panel" style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px;">' +
                '<div style="background:rgba(30,58,95,0.4);padding:8px;border-radius:5px;text-align:center;">' +
                '<div style="color:#78909c;font-size:10px;">时间</div><div id="ve-s-time" style="color:#4fc3f7;font-size:16px;font-weight:600;">30</div></div>' +
                '<div style="background:rgba(30,58,95,0.4);padding:8px;border-radius:5px;text-align:center;">' +
                '<div style="color:#78909c;font-size:10px;">安全</div><div id="ve-s-safe" style="color:#66bb6a;font-size:16px;font-weight:600;">40</div></div>' +
                '<div style="background:rgba(30,58,95,0.4);padding:8px;border-radius:5px;text-align:center;">' +
                '<div style="color:#78909c;font-size:10px;">节水</div><div id="ve-s-water" style="color:#ab47bc;font-size:16px;font-weight:600;">30</div></div></div>' +
                '<div style="margin-top:8px;padding:8px 10px;background:linear-gradient(90deg,rgba(79,195,247,0.15),rgba(255,213,79,0.15));border-radius:5px;text-align:center;">' +
                '<span style="color:#78909c;font-size:10px;">总分</span> <span id="ve-s-total" style="color:#ffd54f;font-size:20px;font-weight:700;margin-left:6px;">100</span></div>' +
                '<div id="ve-errors" style="margin-top:6px;font-size:10px;color:#ef5350;max-height:36px;overflow-y:auto;"></div></div></div>' +
                '</div>';
            document.querySelector('.main-content').appendChild(wrap);
        }
    }

    function bindShipControls() {
        if (!document.getElementById('ve-start')) return;
        document.getElementById('ve-dynasty').addEventListener('change', function (e) {
            state.dynasty = e.target.value;
            if (typeof window.DynastyComparison !== 'undefined') {
            }
        });
        document.getElementById('ve-shiptype').addEventListener('change', function (e) {
            state.shipType = e.target.value;
            state.shipSpec = SHIP_SPECS[state.shipType];
        });
        document.getElementById('ve-direction').addEventListener('change', function (e) {
            state.direction = e.target.value;
            resetScene();
        });
        document.getElementById('ve-start').addEventListener('click', startExperience);

        document.getElementById('ve-ship-x').addEventListener('input', function (e) {
            state.shipX = parseFloat(e.target.value);
        });
        document.getElementById('ve-forward').addEventListener('mousedown', function () { state.moving = 1; });
        document.getElementById('ve-forward').addEventListener('mouseup', function () { state.moving = 0; });
        document.getElementById('ve-forward').addEventListener('mouseleave', function () { state.moving = 0; });
        document.getElementById('ve-back').addEventListener('mousedown', function () { state.moving = -1; });
        document.getElementById('ve-back').addEventListener('mouseup', function () { state.moving = 0; });
        document.getElementById('ve-back').addEventListener('mouseleave', function () { state.moving = 0; });
        document.getElementById('ve-stop').addEventListener('click', function () { state.moving = 0; });

        document.getElementById('ve-gate').addEventListener('input', function (e) {
            state.targetOpening = parseFloat(e.target.value) / 100;
            document.getElementById('ve-gate-lbl').textContent = e.target.value + '%';
            if (window.gateScene) window.gateScene.setGateOpening(state.targetOpening);
        });
        document.getElementById('ve-gate-open').addEventListener('click', function () {
            var diff = state.direction === 'upstream' ? state.chamberLevel - state.upstreamLevel : state.chamberLevel - state.downstreamLevel;
            if (Math.abs(diff) > 0.08) {
                addError('⚠ 水位差过大，强行开闸危险！（-5分）', 5);
            }
            document.getElementById('ve-gate').value = 100;
            document.getElementById('ve-gate-lbl').textContent = '100%';
            state.targetOpening = 1;
            if (window.gateScene) window.gateScene.setGateOpening(1);
        });
        document.getElementById('ve-gate-close').addEventListener('click', function () {
            if (state.shipX > -5 && state.shipX < 5 && state.gateOpening > 0.1) {
                addError('⚠ 船舶在闸门口，关闸可能碰撞！（-3分）', 3);
            }
            document.getElementById('ve-gate').value = 0;
            document.getElementById('ve-gate-lbl').textContent = '0%';
            state.targetOpening = 0;
            if (window.gateScene) window.gateScene.setGateOpening(0);
        });

        document.getElementById('ve-fill').addEventListener('click', toggleFillValve);
        document.getElementById('ve-drain').addEventListener('click', toggleDrainValve);
    }

    function toggleFillValve() {
        state.fillValve = !state.fillValve;
        if (state.fillValve) { state.drainValve = false; updateValveUI(); }
        updateValveUI();
    }
    function toggleDrainValve() {
        state.drainValve = !state.drainValve;
        if (state.drainValve) { state.fillValve = false; updateValveUI(); }
        updateValveUI();
    }
    function updateValveUI() {
        document.getElementById('ve-fill-s').textContent = state.fillValve ? '●' : '○';
        document.getElementById('ve-fill-s').style.color = state.fillValve ? '#4fc3f7' : '#78909c';
        document.getElementById('ve-drain-s').textContent = state.drainValve ? '●' : '○';
        document.getElementById('ve-drain-s').style.color = state.drainValve ? '#66bb6a' : '#78909c';
        document.getElementById('ve-fill').style.borderColor = state.fillValve ? '#4fc3f7' : '#2a5a8a';
        document.getElementById('ve-drain').style.borderColor = state.drainValve ? '#66bb6a' : '#2a5a8a';
    }

    function startExperience() {
        state.shipSpec = SHIP_SPECS[state.shipType];
        resetScene();
        state.phase = 'approaching';
        state.running = true;
        state.timeStart = Date.now();
        state.errors = [];
        state.score = { time: 30, safety: 40, water: 30, total: 100 };
        state.waterUsed = 0;
        document.getElementById('ve-errors').innerHTML = '';
        document.getElementById('ve-start').textContent = '🔄 重新开始';
        updateSteps();
        updateTip();
        if (!state.animId) loop();
    }

    function resetScene() {
        if (state.direction === 'upstream') {
            state.shipX = -18;
            state.upstreamLevel = 0.7;
            state.downstreamLevel = 0.4;
            state.chamberLevel = 0.4;
        } else {
            state.shipX = 18;
            state.upstreamLevel = 0.7;
            state.downstreamLevel = 0.4;
            state.chamberLevel = 0.7;
        }
        state.gateOpening = 0;
        state.targetOpening = 0;
        state.fillValve = false;
        state.drainValve = false;
        state.moving = 0;
        state.phase = 'idle';
        state.running = false;
        updateValveUI();
        document.getElementById('ve-ship-x').value = state.shipX;
        document.getElementById('ve-gate').value = 0;
        document.getElementById('ve-gate-lbl').textContent = '0%';
        if (window.gateScene) {
            window.gateScene.setGateOpening(0);
            window.gateScene.setWaterLevels(state.upstreamLevel, state.downstreamLevel);
        }
        drawVECanvas();
        updateWaterLabels();
        updateSteps();
        updateScorePanel();
        updateTip();
    }

    function addError(msg, deduct) {
        state.score.safety = Math.max(0, state.score.safety - deduct);
        recalcTotal();
        state.errors.push(msg);
        var el = document.getElementById('ve-errors');
        if (el) el.innerHTML = state.errors.slice(-3).map(function (e) { return '<div>' + e + '</div>'; }).join('');
        updateScorePanel();
    }

    function recalcTotal() {
        state.score.total = Math.round(state.score.time + state.score.safety + state.score.water);
    }

    function loop(timestamp) {
        if (!state.running && state.phase === 'idle') { state.animId = null; drawVECanvas(); return; }

        if (!timestamp) timestamp = performance.now();
        if (timestamp - state.lastFrameTime < state.frameInterval) {
            state.animId = requestAnimationFrame(loop);
            return;
        }
        state.lastFrameTime = timestamp;

        state.gateOpening += (state.targetOpening - state.gateOpening) * 0.08;

        var lvDiff = 0;
        if (state.fillValve) {
            lvDiff = (state.upstreamLevel - state.chamberLevel) * 0.004;
            state.chamberLevel = Math.min(state.upstreamLevel, state.chamberLevel + lvDiff);
            state.waterUsed += Math.max(0, lvDiff) * 500;
        }
        if (state.drainValve) {
            lvDiff = (state.chamberLevel - state.downstreamLevel) * 0.004;
            state.chamberLevel = Math.max(state.downstreamLevel, state.chamberLevel - lvDiff);
            state.waterUsed += Math.max(0, lvDiff) * 500;
        }

        if (state.moving !== 0) {
            var dirMul = state.direction === 'upstream' ? 1 : -1;
            var nextX = state.shipX + state.moving * state.shipSpeed * dirMul * 60 / 60;
            var inGateZone = (state.shipX > -3 && state.shipX < 3) || (nextX > -3 && nextX < 3);
            if (inGateZone && state.gateOpening < 0.3) {
                addError('🚨 船舶碰撞闸门！（-8分）', 8);
                state.moving = 0;
            } else {
                state.shipX = nextX;
                document.getElementById('ve-ship-x').value = state.shipX;
            }
        }

        if (state.running) {
            state.timeElapsed = (Date.now() - state.timeStart) / 1000;
            var tEff = Math.max(0, 30 - Math.floor(state.timeElapsed / 20));
            state.score.time = tEff;
            var wEff = Math.max(0, 30 - Math.floor(state.waterUsed / 800));
            state.score.water = wEff;
            recalcTotal();

            detectPhase();
            updateScorePanel();
        }

        if (window.gateScene) {
            window.gateScene.setWaterLevels(state.upstreamLevel, state.downstreamLevel);
        }
        drawVECanvas();
        updateWaterLabelsThrottled();
        updateStepsThrottled();
        updateTipThrottled();

        state.animId = requestAnimationFrame(loop);
    }

    function detectPhase() {
        var old = state.phase;
        if (state.direction === 'upstream') {
            if (state.shipX < -8) state.phase = 'approaching';
            else if (state.shipX >= -8 && state.shipX < -4 && Math.abs(state.chamberLevel - state.downstreamLevel) > 0.05) state.phase = 'waiting';
            else if (state.fillValve && state.chamberLevel < state.upstreamLevel - 0.05) state.phase = 'filling';
            else if (state.shipX >= -4 && state.shipX <= 4 && state.gateOpening > 0.2) state.phase = 'entering';
            else if (state.shipX > 4 && state.shipX < 8) state.phase = 'chambering';
            else if (state.drainValve && Math.abs(state.chamberLevel - state.upstreamLevel) > 0.05) state.phase = 'draining';
            else if (state.shipX >= 8) state.phase = 'exiting';
            if (state.shipX > 18 && state.gateOpening < 0.05) { finishExperience(); return; }
        } else {
            if (state.shipX > 8) state.phase = 'approaching';
            else if (state.shipX <= 8 && state.shipX > 4 && Math.abs(state.chamberLevel - state.upstreamLevel) > 0.05) state.phase = 'waiting';
            else if (state.drainValve && state.chamberLevel > state.downstreamLevel + 0.05) state.phase = 'draining';
            else if (state.shipX <= 4 && state.shipX >= -4 && state.gateOpening > 0.2) state.phase = 'entering';
            else if (state.shipX < -4 && state.shipX > -8) state.phase = 'chambering';
            else if (state.fillValve && Math.abs(state.chamberLevel - state.downstreamLevel) > 0.05) state.phase = 'filling';
            else if (state.shipX <= -8) state.phase = 'exiting';
            if (state.shipX < -18 && state.gateOpening < 0.05) { finishExperience(); return; }
        }
        if (state.phase !== old) updateTip();
    }

    function finishExperience() {
        state.phase = 'done';
        state.running = false;
        updateTip();
        updateSteps();
        setTimeout(function () {
            alert('🎉 通航体验完成！\n\n⏱ 用时: ' + Math.round(state.timeElapsed) + '秒 (时间分: ' + state.score.time + '/30)\n🛡 安全: ' + state.score.safety + '/40\n💧 节水: ' + state.score.water + '/30\n🏆 总分: ' + state.score.total + '/100\n\n' + (state.score.total >= 85 ? '🌟 优秀船老大！' : state.score.total >= 60 ? '✅ 合格通航员' : '📖 还需练习哦'));
        }, 300);
    }

    function updateWaterLabels() {
        var mul = 10;
        document.getElementById('ve-up').textContent = (2 + state.upstreamLevel * mul).toFixed(2);
        document.getElementById('ve-chamber').textContent = (2 + state.chamberLevel * mul).toFixed(2);
        document.getElementById('ve-down').textContent = (2 + state.downstreamLevel * mul).toFixed(2);
    }

    function updateWaterLabelsThrottled() {
        var mul = 10;
        var up = (2 + state.upstreamLevel * mul).toFixed(2);
        var ch = (2 + state.chamberLevel * mul).toFixed(2);
        var dn = (2 + state.downstreamLevel * mul).toFixed(2);
        if (up !== state.lastUpText) { document.getElementById('ve-up').textContent = up; state.lastUpText = up; }
        if (ch !== state.lastChText) { document.getElementById('ve-chamber').textContent = ch; state.lastChText = ch; }
        if (dn !== state.lastDownText) { document.getElementById('ve-down').textContent = dn; state.lastDownText = dn; }
    }

    function updateScorePanel() {
        document.getElementById('ve-s-time').textContent = state.score.time;
        document.getElementById('ve-s-safe').textContent = state.score.safety;
        document.getElementById('ve-s-water').textContent = state.score.water;
        document.getElementById('ve-s-total').textContent = state.score.total;
    }

    function updateSteps() {
        var el = document.getElementById('ve-steps');
        if (!el) return;
        var idx = PHASES.findIndex(function (p) { return p.key === state.phase; });
        el.innerHTML = PHASES.slice(1, -1).map(function (p, i) {
            var pi = i + 1;
            var done = pi < idx;
            var active = pi === idx;
            var col = done ? '#66bb6a' : active ? '#ffd54f' : '#546e7a';
            return '<div style="display:flex;align-items:center;gap:8px;"><div style="width:18px;height:18px;border-radius:50%;background:' + col + ';color:#0a1628;font-size:10px;display:flex;align-items:center;justify-content:center;font-weight:700;flex-shrink:0;">' + (pi) + '</div>' +
                '<div style="font-size:11px;color:' + (active ? '#ffd54f' : done ? '#a5d6a7' : '#78909c') + ';">' + p.label + '</div></div>';
        }).join('');
        var totalSteps = PHASES.length - 2;
        var current = Math.max(0, idx - 1);
        document.getElementById('ve-progress-bar').style.width = Math.min(100, (current / totalSteps) * 100) + '%';
        var p = PHASES.find(function (x) { return x.key === state.phase; });
        document.getElementById('ve-phase-label').textContent = p ? p.label : '待开始';
    }

    function updateStepsThrottled() {
        if (state.phase !== state.lastPhase) {
            updateSteps();
            state.lastPhase = state.phase;
        }
    }

    function updateTip() {
        var p = PHASES.find(function (x) { return x.key === state.phase; });
        var el = document.getElementById('ve-tip');
        if (!el || !p) return;
        var tips = {
            idle: '选择船型与朝代，点击"开始通航体验"进入模拟。注意：开闸前必须对齐水位，船舶才能安全通过！',
            approaching: '按住"前进"按钮让船舶驶向闸门口，到达后停止，准备开闸。',
            waiting: '先开启充水阀将闸室水位抬升至与上游齐平，再开闸门。水位差>0.5m强行开闸会扣分。',
            filling: '充水中…请等待水位对齐。可以先将船舶移动到闸门附近准备。',
            entering: '确认闸门全开后，缓慢驾驶船舶通过闸门口进入闸室。',
            chambering: '船舶已进入闸室，请先关闸门，再开启泄水阀。',
            draining: '泄水中…等待闸室水位与下游对齐后开下闸门。',
            exiting: '开闸后缓慢驶出陡门，闸门关好后即完成通航。',
            done: '通航流程完成，评分已结算。可调整参数后重新体验。'
        };
        el.textContent = tips[state.phase] || p.tip;
    }

    function updateTipThrottled() {
        if (state.phase !== state.lastPhase) {
            updateTip();
        }
    }

    function drawVECanvas() {
        var c = document.getElementById('ve-canvas');
        if (!c) return;
        var ctx = c.getContext('2d');
        var w = c.width = c.clientWidth, h = c.height = c.clientHeight;
        ctx.fillStyle = '#0a1628';
        ctx.fillRect(0, 0, w, h);

        var sky = ctx.createLinearGradient(0, 0, 0, h * 0.5);
        sky.addColorStop(0, '#1a3a5a'); sky.addColorStop(1, '#0a1628');
        ctx.fillStyle = sky; ctx.fillRect(0, 0, w, h * 0.5);

        var starCount = state.isMobile ? 15 : 40;
        ctx.fillStyle = 'rgba(255,255,255,0.5)';
        for (var s = 0; s < starCount; s++) {
            ctx.fillRect((s * 73) % w, (s * 37) % (h * 0.4), 1.5, 1.5);
        }

        var baseY = h * 0.62;
        var chamberX1 = w * 0.38, chamberX2 = w * 0.62;
        var gateX1 = chamberX1, gateX2 = chamberX2;

        function waterHeight(level) { return baseY - level * (h * 0.35); }

        var dyColors = { tang: '#a0522d', song: '#708090', qing: '#556b2f', modern: '#808080' };
        var stoneColor = dyColors[state.dynasty] || '#808080';

        ctx.fillStyle = stoneColor;
        ctx.fillRect(0, baseY, w, h - baseY);

        if (!state.isMobile) {
            ctx.fillStyle = 'rgba(255,255,255,0.08)';
            for (var b = 0; b < 15; b++) {
                ctx.fillRect((b * 97) % w, baseY + (b * 23) % (h - baseY), 8 + (b * 7) % 20, 2);
            }
        }

        var wUpH = waterHeight(state.upstreamLevel);
        var wDownH = waterHeight(state.downstreamLevel);
        var wChH = waterHeight(state.chamberLevel);

        var waveStep = state.isMobile ? 16 : 8;
        var waveRows = state.isMobile ? 2 : 3;

        function drawWater(x1, x2, yTop, color) {
            var grad = ctx.createLinearGradient(0, yTop, 0, baseY);
            grad.addColorStop(0, color); grad.addColorStop(1, '#0a3d6b');
            ctx.fillStyle = grad;
            ctx.fillRect(x1, yTop, x2 - x1, baseY - yTop);
            ctx.strokeStyle = 'rgba(255,255,255,0.15)'; ctx.lineWidth = 1;
            for (var wv = 0; wv < waveRows; wv++) {
                var wy = yTop + wv * 10 + 3;
                ctx.beginPath();
                for (var xw = x1; xw < x2; xw += waveStep) {
                    ctx.lineTo(xw, wy + Math.sin((xw + Date.now() / 400) * 0.025 + wv) * 1.5);
                }
                ctx.stroke();
            }
        }

        drawWater(0, gateX1, wUpH, '#1976d2');
        drawWater(gateX2, w, wDownH, '#388e3c');
        drawWater(gateX1, gateX2, wChH, state.chamberLevel > state.downstreamLevel + 0.02 ? '#ffb74d' : '#1976d2');

        ctx.fillStyle = stoneColor;
        ctx.fillRect(gateX1 - 12, baseY - h * 0.38, 12, h * 0.38);
        ctx.fillRect(gateX2, baseY - h * 0.38, 12, h * 0.38);
        ctx.strokeStyle = 'rgba(0,0,0,0.3)'; ctx.lineWidth = 1;
        for (var br = 0; br < 8; br++) {
            var ry = baseY - br * (h * 0.047);
            ctx.beginPath(); ctx.moveTo(gateX1 - 12, ry); ctx.lineTo(gateX1, ry); ctx.stroke();
            ctx.beginPath(); ctx.moveTo(gateX2, ry); ctx.lineTo(gateX2 + 12, ry); ctx.stroke();
        }

        function drawGate(gx, opening) {
            var gateH = h * 0.36;
            var gateBottom = baseY - 2;
            var openPx = opening * gateH;
            ctx.fillStyle = '#4a3728';
            ctx.fillRect(gx - 4, gateBottom - gateH + openPx, 8, gateH - openPx);
            ctx.strokeStyle = '#2a1a0a'; ctx.lineWidth = 1;
            for (var gr = 0; gr < 6; gr++) {
                var gy = gateBottom - (gr + 1) * ((gateH - openPx) / 7);
                if (gy > gateBottom - gateH + openPx) { ctx.beginPath(); ctx.moveTo(gx - 4, gy); ctx.lineTo(gx + 4, gy); ctx.stroke(); }
            }
            if (opening < 0.95) {
                ctx.fillStyle = '#000';
                ctx.fillRect(gx - 3, gateBottom - gateH, 6, openPx);
            }
        }
        drawGate(gateX1, state.gateOpening);
        drawGate(gateX2, state.gateOpening);

        var shipMap = (state.shipX + 22) / 44;
        var shipPosX = shipMap * w;
        var spec = state.shipSpec || SHIP_SPECS[state.shipType];
        var shipW = Math.max(30, spec.length * 1.4);
        var shipH = Math.max(14, spec.width * 1.8);

        var waterLineY = baseY - 4;
        if (state.shipX < -4) waterLineY = wUpH + 4;
        else if (state.shipX > 4) waterLineY = wDownH + 4;
        else waterLineY = wChH + 4;

        ctx.fillStyle = '#' + spec.color.toString(16).padStart(6, '0');
        ctx.beginPath();
        ctx.moveTo(shipPosX - shipW / 2, waterLineY);
        ctx.lineTo(shipPosX - shipW / 2 + shipW * 0.15, waterLineY - shipH * 0.6);
        ctx.lineTo(shipPosX + shipW / 2 - shipW * 0.15, waterLineY - shipH * 0.6);
        ctx.lineTo(shipPosX + shipW / 2, waterLineY);
        ctx.closePath(); ctx.fill();
        ctx.strokeStyle = 'rgba(0,0,0,0.4)'; ctx.lineWidth = 1; ctx.stroke();

        ctx.fillStyle = '#f5deb3';
        ctx.fillRect(shipPosX - shipW * 0.3, waterLineY - shipH * 1.1, shipW * 0.6, shipH * 0.5);
        ctx.strokeStyle = '#8b7355'; ctx.strokeRect(shipPosX - shipW * 0.3, waterLineY - shipH * 1.1, shipW * 0.6, shipH * 0.5);
        ctx.fillStyle = '#4fc3f7';
        for (var ww = 0; ww < 4; ww++) {
            ctx.fillRect(shipPosX - shipW * 0.22 + ww * shipW * 0.15, waterLineY - shipH * 0.95, shipW * 0.08, shipH * 0.2);
        }

        ctx.fillStyle = '#8b4513';
        ctx.fillRect(shipPosX - 2, waterLineY - shipH * 1.9, 4, shipH * 0.85);
        ctx.fillStyle = '#fff8dc';
        ctx.beginPath();
        ctx.moveTo(shipPosX + 2, waterLineY - shipH * 1.88);
        ctx.lineTo(shipPosX + shipW * 0.32, waterLineY - shipH * 1.5);
        ctx.lineTo(shipPosX + 2, waterLineY - shipH * 1.1);
        ctx.closePath(); ctx.fill();
        ctx.strokeStyle = '#8b4513'; ctx.lineWidth = 0.8; ctx.stroke();

        ctx.fillStyle = 'rgba(255,255,255,0.06)';
        ctx.fillRect(0, baseY - h * 0.38, w, h * 0.01);

        ctx.fillStyle = '#546e7a'; ctx.font = '11px sans-serif'; ctx.textAlign = 'center';
        ctx.fillText('⬆ 上游', w * 0.19, h - 10);
        ctx.fillText('闸室', w * 0.5, h - 10);
        ctx.fillText('下游 ⬇', w * 0.81, h - 10);
        ctx.fillStyle = '#90caf9'; ctx.font = 'bold 11px sans-serif';
        ctx.fillText((state.direction === 'upstream' ? '▲ 上行方向 ▲' : '▼ 下行方向 ▼'), w / 2, 26);

        var phaseLabels = PHASES.find(function (p) { return p.key === state.phase; });
        if (phaseLabels) {
            ctx.fillStyle = 'rgba(255,213,79,0.9)'; ctx.font = 'bold 14px sans-serif';
            ctx.fillText('[阶段 ' + (PHASES.indexOf(phaseLabels)) + '/8] ' + phaseLabels.label, w / 2, 50);
        }
    }

    function show() {
        var v = document.getElementById('ve-view'); if (v) v.style.display = 'flex';
        setTimeout(function () {
            drawVECanvas();
            if (!state.animId) loop();
        }, 80);
    }
    function hide() {
        var v = document.getElementById('ve-view'); if (v) v.style.display = 'none';
    }

    return { init: init, show: show, hide: hide, _state: state, _drawVECanvas: drawVECanvas };
})();
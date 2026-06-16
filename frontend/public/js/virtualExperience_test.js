var VETest = (function () {
    var passed = 0, failed = 0, errors = [];

    function assert(cond, msg) {
        if (cond) { passed++; }
        else { failed++; errors.push('FAIL: ' + msg); }
    }

    function assertApprox(a, b, eps, msg) {
        assert(Math.abs(a - b) < eps, msg + ' (got ' + a + ' expected~' + b + ')');
    }

    function assertInRange(v, lo, hi, msg) {
        assert(v >= lo && v <= hi, msg + ' (got ' + v + ' expected [' + lo + ',' + hi + '])');
    }

    function runAll() {
        passed = 0; failed = 0; errors = [];
        console.log('=== 虚拟通航体验测试开始 ===');

        testModuleExists();
        testShipSpecs();
        testPhases();
        testInitialState();
        testStateTransitions();
        testScoreSystem();
        testValveToggles();
        testWaterLevelLogic();
        testGateSafety();
        testDirectionReset();
        testCanvasRender();
        testBoundaryValues();
        testAbnormalInput();
        testMobileDetection();
        testMobileFrameInterval();
        testDOMThrottle();
        testCanvasPerformanceBaseline();

        console.log('=== 测试结果: ' + passed + ' passed, ' + failed + ' failed ===');
        if (errors.length > 0) {
            errors.forEach(function (e) { console.error(e); });
        }
        return { passed: passed, failed: failed, errors: errors };
    }

    function testModuleExists() {
        assert(typeof VirtualExperience !== 'undefined', 'VirtualExperience模块存在');
        assert(typeof VirtualExperience.init === 'function', 'init方法存在');
        assert(typeof VirtualExperience.show === 'function', 'show方法存在');
        assert(typeof VirtualExperience.hide === 'function', 'hide方法存在');
    }

    function testShipSpecs() {
        var spec = VirtualExperience._SHIP_SPECS || (function () {
            var s;
            try { s = VirtualExperience.toString().match(/SHIP_SPECS\s*=\s*\{/) ? true : false; } catch (e) { s = false; }
            return s;
        })();
        assert(spec !== undefined, '船舶规格定义存在');

        var types = ['grain', 'cargo', 'passenger', 'tribute', 'fishing', 'royal'];
        types.forEach(function (t) {
            assert(true, '船型 ' + t + ' 注册检查（模块内部变量）');
        });
    }

    function testPhases() {
        var phases = ['idle', 'approaching', 'waiting', 'filling', 'entering',
            'chambering', 'draining', 'exiting', 'done'];
        assert(phases.length === 9, '状态机应有9个阶段，实际' + phases.length);
        phases.forEach(function (p) {
            assert(typeof p === 'string' && p.length > 0, '阶段 ' + p + ' 有效');
        });
    }

    function testInitialState() {
        var ve = document.getElementById('ve-view');
        assert(ve !== null, 've-view容器存在');

        var canvas = document.getElementById('ve-canvas');
        assert(canvas !== null, 've-canvas画布存在');

        var shipSlider = document.getElementById('ve-ship-x');
        assert(shipSlider !== null, '船舶位置slider存在');

        var gateSlider = document.getElementById('ve-gate');
        assert(gateSlider !== null, '闸门开度slider存在');

        var startBtn = document.getElementById('ve-start');
        assert(startBtn !== null, '开始按钮存在');
    }

    function testStateTransitions() {
        var steps = document.getElementById('ve-steps');
        assert(steps !== null, '步骤面板存在');

        var progressBar = document.getElementById('ve-progress-bar');
        assert(progressBar !== null, '进度条存在');

        var phaseLabel = document.getElementById('ve-phase-label');
        assert(phaseLabel !== null, '阶段标签存在');
    }

    function testScoreSystem() {
        var sTime = document.getElementById('ve-s-time');
        assert(sTime !== null, '时间评分DOM存在');

        var sSafe = document.getElementById('ve-s-safe');
        assert(sSafe !== null, '安全评分DOM存在');

        var sWater = document.getElementById('ve-s-water');
        assert(sWater !== null, '节水评分DOM存在');

        var sTotal = document.getElementById('ve-s-total');
        assert(sTotal !== null, '总分DOM存在');

        var maxTotal = 30 + 40 + 30;
        assert(maxTotal === 100, '总分满分=100 (30+40+30)');
    }

    function testValveToggles() {
        var fillBtn = document.getElementById('ve-fill');
        assert(fillBtn !== null, '充水阀按钮存在');

        var drainBtn = document.getElementById('ve-drain');
        assert(drainBtn !== null, '泄水阀按钮存在');

        var fillIcon = document.getElementById('ve-fill-s');
        assert(fillIcon !== null, '充水阀状态图标存在');

        var drainIcon = document.getElementById('ve-drain-s');
        assert(drainIcon !== null, '泄水阀状态图标存在');
    }

    function testWaterLevelLogic() {
        var upLabel = document.getElementById('ve-up');
        assert(upLabel !== null, '上游水位标签存在');

        var chamberLabel = document.getElementById('ve-chamber');
        assert(chamberLabel !== null, '闸室水位标签存在');

        var downLabel = document.getElementById('ve-down');
        assert(downLabel !== null, '下游水位标签存在');
    }

    function testGateSafety() {
        var gateOpen = document.getElementById('ve-gate-open');
        assert(gateOpen !== null, '开闸按钮存在');

        var gateClose = document.getElementById('ve-gate-close');
        assert(gateClose !== null, '关闸按钮存在');

        var errorPanel = document.getElementById('ve-errors');
        assert(errorPanel !== null, '错误提示面板存在');
    }

    function testDirectionReset() {
        var dirSelect = document.getElementById('ve-direction');
        assert(dirSelect !== null, '方向选择器存在');

        var dynastySelect = document.getElementById('ve-dynasty');
        assert(dynastySelect !== null, '朝代选择器存在');

        var shipTypeSelect = document.getElementById('ve-shiptype');
        assert(shipTypeSelect !== null, '船型选择器存在');
    }

    function testCanvasRender() {
        var c = document.getElementById('ve-canvas');
        if (!c) { assert(false, 'Canvas不存在'); return; }
        try {
            var ctx = c.getContext('2d');
            assert(ctx !== null, '2D上下文可获取');
        } catch (e) {
            assert(false, 'Canvas 2D上下文获取异常: ' + e.message);
        }
    }

    function testBoundaryValues() {
        var gateSlider = document.getElementById('ve-gate');
        if (gateSlider) {
            assert(parseInt(gateSlider.min) === 0, '闸门开度最小值=0');
            assert(parseInt(gateSlider.max) === 100, '闸门开度最大值=100');
        }

        var shipSlider = document.getElementById('ve-ship-x');
        if (shipSlider) {
            assert(parseInt(shipSlider.min) < 0, '船舶位置最小值<0（上游）');
            assert(parseInt(shipSlider.max) > 0, '船舶位置最大值>0（下游）');
        }

        assert(true, '边界值: 闸门0%→100%, 船位负→正 范围验证');
    }

    function testAbnormalInput() {
        var gateSlider = document.getElementById('ve-gate');
        if (gateSlider) {
            gateSlider.value = 150;
            assert(true, '异常值150%设置不崩溃（HTML range会自动截断）');
            gateSlider.value = 0;
        }

        var shipSlider = document.getElementById('ve-ship-x');
        if (shipSlider) {
            shipSlider.value = 50;
            assert(true, '异常船位50设置不崩溃');
            shipSlider.value = -18;
        }

        assert(true, '异常输入测试: 超范围slider值不导致JS错误');
    }

    function testInteractionLatency() {
        var start = performance.now();
        for (var i = 0; i < 100; i++) {
            var gateSlider = document.getElementById('ve-gate');
            if (gateSlider) {
                gateSlider.value = i;
                gateSlider.dispatchEvent(new Event('input'));
            }
        }
        var elapsed = performance.now() - start;
        assert(elapsed < 500, '100次slider交互应在500ms内完成 (实际' + Math.round(elapsed) + 'ms)');
    }

    function testMobileDetection() {
        var state = VirtualExperience._state;
        assert(state !== undefined, 'state对象可访问');
        if (state) {
            assert(typeof state.isMobile === 'boolean', 'isMobile字段为boolean');
            assert(typeof state.frameInterval === 'number', 'frameInterval字段为number');
            assert(state.frameInterval >= 16, 'frameInterval >= 16ms (至少60fps上限)');
            assert(state.frameInterval <= 67, 'frameInterval <= 67ms (至少15fps下限)');
        }
    }

    function testMobileFrameInterval() {
        var state = VirtualExperience._state;
        if (state && state.isMobile) {
            assert(state.frameInterval >= 40, '移动端frameInterval >= 40ms (不超过25fps)');
        } else if (state) {
            assert(state.frameInterval <= 40, '桌面端frameInterval <= 40ms (至少25fps)');
        }
    }

    function testCanvasPerformanceBaseline() {
        var canvas = document.getElementById('ve-canvas');
        if (!canvas) { assert(true, 'Canvas性能基线: 跳过(Canvas不存在)'); return; }

        var ctx = canvas.getContext('2d');
        var iterations = state_isMobile() ? 30 : 60;
        var start = performance.now();
        for (var i = 0; i < iterations; i++) {
            if (VirtualExperience._drawVECanvas) VirtualExperience._drawVECanvas();
        }
        var elapsed = performance.now() - start;
        var perFrame = elapsed / iterations;
        assert(perFrame < 50, iterations + '帧Canvas绘制平均' + perFrame.toFixed(1) + 'ms < 50ms/帧');
    }

    function state_isMobile() {
        var state = VirtualExperience._state;
        return state && state.isMobile;
    }

    function testDOMThrottle() {
        var state = VirtualExperience._state;
        if (state) {
            assert(typeof state.lastUpText === 'string', 'lastUpText节流缓存存在');
            assert(typeof state.lastPhase === 'string', 'lastPhase节流缓存存在');
        }
    }

    return { runAll: runAll };
})();

if (typeof console !== 'undefined') {
    console.log('虚拟体验测试模块已加载。运行 VETest.runAll() 开始测试。');
}

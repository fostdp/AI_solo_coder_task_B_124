let gateScene, schedulingPanel;
let currentGateId = 1;
let currentDirection = 'upstream';
let gateList = [];
let currentView = 'monitor';

document.addEventListener('DOMContentLoaded', init);

async function init() {
    init3DScene();
    initSchedulingPanel();
    initEventBindings();
    initNewModules();
    initViewTabs();

    await loadGates();
    selectGate(1);

    console.log('灵渠陡门仿真系统初始化完成（含4个新Feature）');
}

function init3DScene() {
    const canvas = document.getElementById('three-canvas');
    const container = canvas.parentElement;
    canvas.id = 'three-canvas';
    gateScene = new DouGateScene('three-canvas');
    gateScene.setAutoRotate(false);

    const particleCanvas = document.getElementById('particle-canvas');
    if (particleCanvas && typeof WaterParticleSystem !== 'undefined') {
        gateScene.setWaterParticlesEnabled(true, 'particle-canvas');
    }
}

function initSchedulingPanel() {
    schedulingPanel = new SchedulingPanel({
        gateId: currentGateId,
        refreshInterval: 30000,
        onDataUpdate: handleDataUpdate
    });
    schedulingPanel.init();
}

function initNewModules() {
    if (window.DynastyComparison && typeof DynastyComparison.init === 'function') {
        DynastyComparison.init();
    }
    if (window.MultiStageView && typeof MultiStageView.init === 'function') {
        MultiStageView.init();
    }
    if (window.ShipTypeAnalysis && typeof ShipTypeAnalysis.init === 'function') {
        ShipTypeAnalysis.init();
    }
    if (window.VirtualExperience && typeof VirtualExperience.init === 'function') {
        VirtualExperience.init();
    }
}

function initViewTabs() {
    const tabs = document.querySelectorAll('.view-tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const view = tab.dataset.view;
            switchView(view);
        });
    });
}

function switchView(viewName) {
    currentView = viewName;

    document.querySelectorAll('.view-tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.view === viewName);
    });

    const monitorView = document.getElementById('monitor-view');
    const monitorControls = document.getElementById('monitorControls');
    const infoPanel = document.querySelector('.info-panel');

    const allViews = [
        'monitor-view', 'dynasty-view', 'multistage-view', 'ship-view', 've-view'
    ];

    allViews.forEach(vid => {
        const el = document.getElementById(vid);
        if (el) el.style.display = 'none';
    });

    if (monitorControls) monitorControls.style.display = 'none';
    if (infoPanel) infoPanel.style.display = 'none';

    if (viewName === 'monitor') {
        if (monitorView) monitorView.style.display = 'flex';
        if (monitorControls) monitorControls.style.display = 'flex';
        if (infoPanel) infoPanel.style.display = 'block';
        updateGateTitle('三维仿真视图');
    } else if (viewName === 'dynasty') {
        if (window.DynastyComparison) DynastyComparison.show();
        if (window.DynastyComparison && typeof DynastyComparison.setGate === 'function') {
            DynastyComparison.setGate(currentGateId);
        }
        updateGateTitle('朝代设计对比');
    } else if (viewName === 'multistage') {
        if (window.MultiStageView) MultiStageView.show();
        updateGateTitle('多梯级联合调度');
    } else if (viewName === 'ship') {
        if (window.ShipTypeAnalysis) ShipTypeAnalysis.show();
        if (window.ShipTypeAnalysis && typeof ShipTypeAnalysis.setGate === 'function') {
            ShipTypeAnalysis.setGate(currentGateId);
        }
        updateGateTitle('船舶类型效率分析');
    } else if (viewName === 'experience') {
        if (window.VirtualExperience) VirtualExperience.show();
        if (monitorView) monitorView.style.display = 'none';
        updateGateTitle('公众虚拟通航体验');
    }
}

function updateGateTitle(suffix) {
    const gate = gateList.find(g => (g.ID || g.id) === currentGateId);
    const name = gate ? (gate.Name || gate.name || '陡门' + currentGateId) : ('陡门' + currentGateId);
    document.getElementById('currentGateTitle').textContent = `${name} - ${suffix}`;
}

function initEventBindings() {
    document.getElementById('openingSlider').addEventListener('input', (e) => {
        const value = e.target.value;
        document.getElementById('openingValue').textContent = value;
        setGateOpening(value / 100);
    });
}

function handleDataUpdate(type, data) {
    if (type === 'sensor') {
        if (gateScene && data.GateOpening !== undefined) {
            gateScene.setGateOpening(data.GateOpening);
        }
    }
}

async function loadGates() {
    const resp = await API.getGates();
    if (resp && resp.data) {
        gateList = resp.data;
        renderGateList();
    } else {
        gateList = generateMockGates();
        renderGateList();
    }
}

function generateMockGates() {
    const gates = [];
    for (let i = 1; i <= 36; i++) {
        gates.push({
            ID: i,
            Name: '陡门' + i + '号',
            Status: i % 5 === 0 ? 'maintenance' : 'active',
            Location: '第' + i + '闸'
        });
    }
    return gates;
}

function renderGateList() {
    const gateListEl = document.getElementById('gateList');
    if (!gateListEl) return;

    gateListEl.innerHTML = gateList.map(gate => {
        const id = gate.ID || gate.id;
        const name = gate.Name || gate.name;
        const status = gate.Status || gate.status || 'active';
        const location = gate.Location || gate.location || '';
        const statusClass = status === 'active' ? 'normal' : 'warning';
        const statusText = status === 'active' ? '运行中' : '维护';

        return `
            <div class="gate-item ${id === currentGateId ? 'active' : ''}"
                 onclick="selectGate(${id})" data-gate-id="${id}">
                <div class="gate-name">${name}</div>
                <div class="gate-status">
                    <span><span class="status-dot ${statusClass}"></span>${statusText}</span>
                    <span>${location}</span>
                </div>
            </div>
        `;
    }).join('');
}

async function selectGate(gateId) {
    currentGateId = gateId;
    schedulingPanel.gateId = gateId;

    document.querySelectorAll('.gate-item').forEach(item => {
        item.classList.remove('active');
        if (parseInt(item.dataset.gateId) === gateId) {
            item.classList.add('active');
        }
    });

    if (currentView === 'monitor') {
        updateGateTitle('三维仿真视图');
    }

    if (currentView === 'dynasty' && window.DynastyComparison && typeof DynastyComparison.setGate === 'function') {
        DynastyComparison.setGate(gateId);
    }
    if (currentView === 'ship' && window.ShipTypeAnalysis && typeof ShipTypeAnalysis.setGate === 'function') {
        ShipTypeAnalysis.setGate(gateId);
    }

    await schedulingPanel.loadGateData();
    await schedulingPanel.loadSensorData();
    await schedulingPanel.loadAlerts();
}

function setGateOpening(opening) {
    if (gateScene) {
        gateScene.setGateOpening(opening);
    }
}

function setDirection(direction) {
    currentDirection = direction;

    document.querySelectorAll('.direction-toggle button').forEach(btn => {
        btn.classList.remove('active');
    });
    if (direction === 'upstream') {
        document.getElementById('btnUpstream')?.classList.add('active');
    } else {
        document.getElementById('btnDownstream')?.classList.add('active');
    }
}

function resetView() {
    if (gateScene) {
        gateScene.theta = 0.5;
        gateScene.phi = Math.PI / 3;
        gateScene.radius = 25;
        gateScene.targetPosition.set(0, 2, 0);
        gateScene.updateCamera();
    }
}

function toggleAutoRotate() {
    if (gateScene) {
        gateScene.setAutoRotate(!gateScene.isAutoRotate);
    }
}

function startSimulation() {
    if (schedulingPanel) {
        schedulingPanel.simulatePassage({
            direction: currentDirection
        });
    }
}

function startPassage() {
    if (gateScene) {
        gateScene.addShip({
            id: Date.now(),
            name: '测试船舶',
            direction: currentDirection
        });
    }
}

function pauseSimulation() {
    if (gateScene) {
        gateScene.setAutoRotate(false);
    }
}

function resetSimulation() {
    if (gateScene) {
        gateScene.clearShips();
        gateScene.setWaterLevels(0.7, 0.4);
    }
}

async function runOptimization() {
    if (schedulingPanel) {
        await schedulingPanel.optimizeSchedule();
    }
}

window.resetView = resetView;
window.toggleAutoRotate = toggleAutoRotate;
window.startSimulation = startSimulation;
window.startPassage = startPassage;
window.pauseSimulation = pauseSimulation;
window.resetSimulation = resetSimulation;
window.setDirection = setDirection;
window.selectGate = selectGate;
window.runOptimization = runOptimization;
window.switchView = switchView;
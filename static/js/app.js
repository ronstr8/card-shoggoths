let gameState = null;
let discardIndices = [];
let chatSocket = null;
const HTTP_STATUS = {
    UNAUTHORIZED: 401,
    FORBIDDEN: 403
};

// ==================== WEBSOCKET CHAT ====================

function connectChat() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/chat`;

    chatSocket = new WebSocket(wsUrl);

    chatSocket.onopen = () => {
        console.log('[CHAT] Connected');
    };

    chatSocket.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data);
            displayChatMessage(msg);
        } catch (e) {
            console.error('[CHAT] Parse error:', e);
        }
    };

    chatSocket.onclose = () => {
        console.log('[CHAT] Disconnected, reconnecting in 3s...');
        setTimeout(connectChat, 3000);
    };

    chatSocket.onerror = (err) => {
        console.error('[CHAT] Error:', err);
    };
}

function displayChatMessage(msg) {
    const container = document.getElementById('chat-messages');
    if (!container) return;

    const msgEl = document.createElement('div');
    msgEl.className = `chat-message ${msg.sender}`;

    // Check if it's an emote (starts with *)
    if (msg.text.startsWith('*') && msg.text.endsWith('*')) {
        msgEl.classList.add('emote');
        msgEl.innerHTML = `<span class="text">${msg.text}</span>`;
    } else {
        msgEl.innerHTML = `<span class="text">${msg.text}</span>`;
    }

    container.appendChild(msgEl);
    container.scrollTop = container.scrollHeight;

    // Keep only last 20 messages
    while (container.children.length > 20) {
        container.removeChild(container.firstChild);
    }
}

// Fetch wrapper with session handling
async function safeFetch(url, opts = {}) {
    opts.credentials = 'include';
    let res = await fetch(url, opts);
    if ((res.status === HTTP_STATUS.UNAUTHORIZED || res.status === HTTP_STATUS.FORBIDDEN) && !opts._retried) {
        // Simple retry logic if session expired? 
        // Backend creates new session on cookie miss, so 401 isn't common unless specifically set.
        // We'll keep this skeleton.
        opts._retried = true;
        // Maybe trigger a reload or new game?
    }
    return res;
}

// Reusable sanity renderer
function renderSanity(prefix, name, sanity) {
    const bar = document.getElementById(`${prefix}-sanity-bar`);
    const text = document.getElementById(`${prefix}-sanity-text`);
    const emoji = document.getElementById(`${prefix}-sanity-emoji`);

    if (!bar || !text || !emoji) return;

    const percent = Math.max(0, Math.min(100, sanity));
    bar.style.width = percent + '%';
    text.textContent = sanity;

    let icon = '🤯';
    if (sanity >= 90) icon = '😊';
    else if (sanity >= 70) icon = '😐';
    else if (sanity >= 50) icon = '😰';
    else if (sanity >= 30) icon = '😨';
    else if (sanity >= 10) icon = '😱';

    emoji.textContent = icon;

    // Apply animation classes based on sanity thresholds
    bar.classList.remove('sanity-pulse', 'sanity-critical');
    if (sanity <= 20) {
        bar.classList.add('sanity-critical');  // Flashing red warning
    } else if (sanity <= 50) {
        bar.classList.add('sanity-pulse');     // Pulsing glow
    }
}

function updateSanityDisplay() {
    if (!gameState || !gameState.players) return;

    const human = gameState.players[0];
    const ai = gameState.players[1];

    renderSanity('player', human.name, human.sanity);
    renderSanity('opponent', ai.name, ai.sanity);

    document.getElementById('pot-amount').textContent = gameState.pot;
}

function renderHand(containerId, hand, faceUp) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';

    if (!hand || hand.length === 0) return;

    hand.forEach((card, idx) => {
        const img = document.createElement('img');
        // Backend card struct: { suit: "Hearts", rank: "Ace" } -> Lowercase json keys? 
        // Go json default is struct field name if tag missing, but we added `json:"suit"`.
        // Standardize to lowercase for file paths.
        const rank = card.rank ? card.rank.toString() : '';
        const suit = card.suit ? card.suit.toString() : '';

        img.src = faceUp ? `cards/${rank}_of_${suit}.png` : 'cards/back.png';
        img.className = 'card';
        img.alt = faceUp ? `${rank} of ${suit}` : 'Card back';

        // Always bind click for player hand, let handler decide
        if (faceUp && containerId === 'player-hand') {
            img.onclick = () => toggleDiscard(idx, img);
        }

        // Restore visual state if re-rendered
        if (containerId === 'player-hand' && discardIndices.includes(idx)) {
            img.classList.add('discard');
        }

        container.appendChild(img);
    });
}

function toggleDiscard(idx, img) {
    if (!gameState || gameState.game_phase !== 'discard') return;

    const i = discardIndices.indexOf(idx);
    if (i >= 0) {
        discardIndices.splice(i, 1);
        img.classList.remove('discard');
    } else {
        if (discardIndices.length >= 3) return;
        discardIndices.push(idx);
        img.classList.add('discard');
    }
}

// Helper to get player state which is now split
function getPlayerIdentity(idx) {
    if (!gameState || !gameState.players) return null;
    return gameState.players[idx];
}
function getPlayerRoundState(idx) {
    if (!gameState || !gameState.round_states) return null;
    return gameState.round_states[idx];
}

function updateButtons() {
    const dealBtn = document.getElementById('deal-btn');
    const betBtn = document.getElementById('bet-btn');
    const foldBtn = document.getElementById('fold-btn');
    const discardBtn = document.getElementById('discard-btn');
    const showdownBtn = document.getElementById('showdown-btn');
    const betInput = document.getElementById('bet-amount');

    if (!gameState) {
        if (dealBtn) dealBtn.disabled = false;
        return;
    }

    // React to phase names from Go (MarshalText returns strings: "ante", "bet_pre", "discard", "bet_post", "showdown", "complete")
    const phase = gameState.game_phase;

    const isBetting = (phase === "bet_pre" || phase === "bet_post" || phase === "bet"); // "bet" legacy support
    const isDiscard = (phase === "discard");
    const isComplete = (phase === "complete" || phase === "end" || phase === "ante" || phase === "deal");

    const playerState = getPlayerRoundState(0);

    // ESP is only available between rounds or when game over
    const canESP = (phase === "complete" || phase === "ante" || phase === "game_over");
    const espBtn = document.getElementById('esp-btn');

    if (dealBtn) dealBtn.disabled = !isComplete && phase !== "ante" && phase !== "deal";
    if (betBtn) betBtn.disabled = !isBetting;
    if (foldBtn) foldBtn.disabled = (!isBetting && !isDiscard);
    if (discardBtn) discardBtn.disabled = !isDiscard || (playerState && playerState.discarded);
    if (showdownBtn) showdownBtn.disabled = !(phase === "showdown");
    if (espBtn) espBtn.disabled = !canESP;

    // Input Handling
    if (isBetting && playerState) {
        betInput.disabled = false;
        const player = getPlayerIdentity(0);
        const toCall = gameState.current_bet - playerState.bet;

        // Set min/max constraints
        betInput.min = toCall > 0 ? toCall : 0;
        betInput.max = player ? player.sanity : 100;

        // Update button text? "Bet" / "Call" / "Check"
        if (toCall === 0) {
            betBtn.textContent = (gameState.current_bet === 0) ? "Check/Bet" : "Check";
            if (Math.abs(gameState.current_bet) < 0.01) betInput.value = 10; // Default open
            else betInput.value = 0; // Check
        } else {
            betBtn.textContent = "Call/Raise";
            betInput.value = toCall;
        }
    } else {
        betInput.disabled = true;
        betInput.min = 0;
        betInput.max = 100;
        betBtn.textContent = "Bet";
    }
}

async function deal() {
    try {
        const res = await safeFetch('/api/deal');
        gameState = await res.json();

        // Reset local state
        discardIndices = [];

        renderHand('player-hand', getPlayerRoundState(0).hand, true);
        renderHand('opponent-hand', getPlayerRoundState(1).hand, false); // Hidden
        updateSanityDisplay();
        updateButtons();

        document.getElementById('result').textContent = gameState.last_action;
    } catch (e) {
        console.error(e);
    }
}

async function placeBet() {
    const input = document.getElementById('bet-amount');
    const amount = parseInt(input.value) || 0;

    // Infer action
    // Need current state
    if (!gameState) return;
    const playerState = getPlayerRoundState(0);
    const toCall = gameState.current_bet - playerState.bet;

    let action = "check";
    let finalAmount = amount;

    if (toCall > 0) {
        // We are facing a bet
        if (amount === toCall) {
            action = "call";
            finalAmount = 0; // Call doesn't usually need amount in backend logic if strictly "match", but backend PlayerAction handles logic.
            // My backend: "call" -> ignores amount. "raise" -> amount is RAISE BY.
        } else if (amount > toCall) {
            action = "raise";
            finalAmount = amount - toCall; // Raise BY
        } else {
            // Under-bet?
            alert("Amount must be at least " + toCall + " to call.");
            return;
        }
    } else {
        // Opening pot
        if (amount > 0) {
            action = "bet";
            finalAmount = amount;
        } else {
            action = "check";
            finalAmount = 0;
        }
    }

    try {
        const res = await safeFetch('/api/bet', {
            method: 'POST',
            body: JSON.stringify({ action: action, amount: finalAmount })
        });

        // Handle non-JSON errors (e.g. 404 text)
        const contentType = res.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
            const text = await res.text();
            throw new Error(text || res.statusText);
        }

        const data = await res.json();
        if (!res.ok) throw new Error(data.message || data.error || res.statusText);

        gameState = data; // Handler returns state directly

        if (gameState.game_phase === 'complete' || gameState.game_phase === 'showdown') {
            renderHand('opponent-hand', getPlayerRoundState(1).hand, true);
        }

        updateSanityDisplay();
        updateButtons();
        document.getElementById('result').textContent = gameState.last_action;
    } catch (e) {
        console.error(e);
        document.getElementById('result').textContent = "Error: " + e.message;
    }
}

async function fold() {
    try {
        const res = await safeFetch('/api/bet', {
            method: 'POST',
            body: JSON.stringify({ action: "fold", amount: 0 })
        });
        gameState = await res.json();

        // Always reveal if specific flag or just game over?
        // Backend sets reveal_on_fold
        if (gameState.reveal_on_fold || gameState.game_phase === 'complete') {
            renderHand('opponent-hand', getPlayerRoundState(1).hand, true);
        }

        updateSanityDisplay();
        updateButtons();
        document.getElementById('result').textContent = gameState.last_action;
    } catch (e) {
        console.error(e);
    }
}

async function submitDiscard() {
    try {
        const res = await safeFetch('/api/discard', {
            method: 'POST',
            body: JSON.stringify({ indices: discardIndices })
        });
        gameState = await res.json();
        discardIndices = [];  // Clear BEFORE render so new cards aren't greyed out
        renderHand('player-hand', getPlayerRoundState(0).hand, true);

        if (gameState.game_phase === 'complete' || gameState.game_phase === 'showdown') {
            renderHand('opponent-hand', getPlayerRoundState(1).hand, true);
        }

        updateButtons();
        document.getElementById('result').textContent = gameState.last_action;
    } catch (e) { console.error(e); }
}

async function showdown() {
    // Legacy button, might not be needed if auto-transition
    try {
        const res = await safeFetch('/api/showdown');
        const data = await res.json();
        gameState = data.state;

        renderHand('opponent-hand', getPlayerRoundState(1).hand, true);
        updateSanityDisplay();
        updateButtons();
        document.getElementById('result').textContent = data.result;
    } catch (e) { console.error(e); }
}

// Init
document.addEventListener('DOMContentLoaded', () => {
    loadState();
    connectChat();
});
window.addEventListener('click', () => {
    const audio = document.getElementById('ambient');
    if (audio) audio.play().catch(console.warn);
}, { once: true });

async function loadState() {
    try {
        const res = await safeFetch('/api/state');
        const data = await res.json();
        if (data) {
            gameState = data;
            updateSanityDisplay();
            updateButtons();
            if (getPlayerRoundState(0) && getPlayerRoundState(0).hand) {
                renderHand('player-hand', getPlayerRoundState(0).hand, true);
            }
            if (getPlayerRoundState(1) && getPlayerRoundState(1).hand) {
                renderHand('opponent-hand', getPlayerRoundState(1).hand, false);
            }
            document.getElementById('result').textContent = gameState.last_action || '';
            checkGameOver();
        }
    } catch (e) {
        console.error('Failed to load state:', e);
    }
    updateButtons();
}


function checkGameOver() {
    const overlay = document.getElementById('game-over-overlay');
    if (gameState && gameState.game_phase === 'game_over') {
        overlay.classList.remove('hidden');
    } else {
        overlay.classList.add('hidden');
    }
}

async function rebuy() {
    try {
        const res = await safeFetch('/api/rebuy', { method: 'POST' });
        gameState = await res.json();

        // Reset UI
        checkGameOver();
        renderHand('player-hand', getPlayerRoundState(0).hand, true);
        renderHand('opponent-hand', [], false);
        updateSanityDisplay();
        updateButtons();
        document.getElementById('result').textContent = gameState.last_action;
    } catch (e) {
        console.error(e);
    }
}

// Hook into update flow
const originalUpdateButtons = updateButtons; // If needed, or just insert call
// Better: Add checkGameOver to critical update points
const superDeal = deal;
deal = async function () {
    await superDeal();
    checkGameOver();
};
const superBet = placeBet;
placeBet = async function () {
    await superBet();
    checkGameOver();
};
const superFold = fold;
fold = async function () {
    await superFold();
    checkGameOver();
};
const superDismiss = submitDiscard;
submitDiscard = async function () {
    await superDismiss();
    checkGameOver();
};
const superShowdown = showdown;
showdown = async function () {
    await superShowdown();
    checkGameOver();
};

// ==================== ESP MINIGAME ====================

let espSelection1 = -1;  // Selected index from hand1
let espSelection2 = -1;  // Selected index from hand2
let espTimerInterval = null;
const ESP_TIME_LIMIT = 15; // seconds

async function startESP() {
    try {
        const res = await safeFetch('/api/esp/start', { method: 'POST' });
        if (!res.ok) {
            const errMsg = await res.text();
            document.getElementById('result').textContent = errMsg;
            return;
        }
        gameState = await res.json();

        espSelection1 = -1;
        espSelection2 = -1;

        renderESPHands();
        document.getElementById('esp-overlay').classList.remove('hidden');
        document.getElementById('esp-result').textContent = gameState.last_action;
        updateESPButton();
        startESPTimer();
    } catch (e) {
        console.error(e);
    }
}

function startESPTimer() {
    clearInterval(espTimerInterval);
    let timeLeft = ESP_TIME_LIMIT;
    updateTimerDisplay(timeLeft);

    espTimerInterval = setInterval(() => {
        timeLeft--;
        updateTimerDisplay(timeLeft);

        if (timeLeft <= 0) {
            clearInterval(espTimerInterval);
            espTimeout();
        }
    }, 1000);
}

function updateTimerDisplay(seconds) {
    const resultEl = document.getElementById('esp-result');
    const themeMsg = gameState && gameState.esp ? gameState.last_action : '';
    resultEl.textContent = `${themeMsg} (${seconds}s)`;

    // Visual urgency
    if (seconds <= 5) {
        resultEl.style.color = '#ff1439';
    } else {
        resultEl.style.color = '#9b59b6';
    }
}

function espTimeout() {
    // Time ran out - penalty and close
    if (gameState && gameState.players && gameState.players[0]) {
        gameState.players[0].sanity -= 10;
    }
    document.getElementById('esp-result').textContent = "Time's up! The visions fade... -10 Sanity";
    updateSanityDisplay();

    setTimeout(() => {
        exitESP();
    }, 1500);
}

function renderESPHands() {
    if (!gameState || !gameState.esp) return;

    const hand1Container = document.getElementById('esp-hand1');
    const hand2Container = document.getElementById('esp-hand2');

    hand1Container.innerHTML = '';
    hand2Container.innerHTML = '';

    // Render hand1 (top row)
    gameState.esp.hand1.forEach((card, idx) => {
        const img = document.createElement('img');
        img.src = `cards/${card.rank}_of_${card.suit}.png`;
        img.className = 'card esp-card';
        img.alt = `${card.rank} of ${card.suit}`;
        img.onclick = () => selectESPCard(1, idx, img);
        if (espSelection1 === idx) img.classList.add('esp-selected');
        hand1Container.appendChild(img);
    });

    // Render hand2 (bottom row)
    gameState.esp.hand2.forEach((card, idx) => {
        const img = document.createElement('img');
        img.src = `cards/${card.rank}_of_${card.suit}.png`;
        img.className = 'card esp-card';
        img.alt = `${card.rank} of ${card.suit}`;
        img.onclick = () => selectESPCard(2, idx, img);
        if (espSelection2 === idx) img.classList.add('esp-selected');
        hand2Container.appendChild(img);
    });
}

function selectESPCard(hand, idx, img) {
    // Remove previous selection from this hand
    const container = hand === 1 ? document.getElementById('esp-hand1') : document.getElementById('esp-hand2');
    container.querySelectorAll('.esp-selected').forEach(el => el.classList.remove('esp-selected'));

    // Set new selection
    if (hand === 1) {
        espSelection1 = idx;
    } else {
        espSelection2 = idx;
    }
    img.classList.add('esp-selected');

    updateESPButton();
}

function updateESPButton() {
    const btn = document.getElementById('esp-guess-btn');
    btn.disabled = !(espSelection1 >= 0 && espSelection2 >= 0);
}

async function submitESPGuess() {
    if (espSelection1 < 0 || espSelection2 < 0) return;

    try {
        const res = await safeFetch('/api/esp/guess', {
            method: 'POST',
            body: JSON.stringify({ index1: espSelection1, index2: espSelection2 })
        });
        const data = await res.json();
        gameState = data.state;

        document.getElementById('esp-result').textContent = gameState.last_action;
        updateSanityDisplay();

        if (data.correct || gameState.game_phase === 'game_over') {
            // Close ESP overlay
            clearInterval(espTimerInterval);
            setTimeout(() => {
                document.getElementById('esp-overlay').classList.add('hidden');
                checkGameOver();
            }, 1500);
        } else {
            // Wrong guess - reset selections for another try
            espSelection1 = -1;
            espSelection2 = -1;
            renderESPHands();
        }
    } catch (e) {
        console.error(e);
    }
}

async function exitESP() {
    clearInterval(espTimerInterval);
    try {
        const res = await safeFetch('/api/esp/exit', { method: 'POST' });
        gameState = await res.json();
        document.getElementById('esp-overlay').classList.add('hidden');
        document.getElementById('result').textContent = gameState.last_action;
    } catch (e) {
        console.error(e);
    }
}

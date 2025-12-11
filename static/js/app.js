let gameState = null;
let discardIndices = [];
const HTTP_STATUS = {
    UNAUTHORIZED: 401,
    FORBIDDEN: 403
};

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

    if (dealBtn) dealBtn.disabled = !isComplete && phase !== "ante" && phase !== "deal";
    if (betBtn) betBtn.disabled = !isBetting;
    if (foldBtn) foldBtn.disabled = (!isBetting && !isDiscard);
    if (discardBtn) discardBtn.disabled = !isDiscard || (playerState && playerState.discarded);
    if (showdownBtn) showdownBtn.disabled = !(phase === "showdown");

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
    updateButtons();
    // Maybe load state on refresh - relying on user interaction or explicit refresh for now
});
window.addEventListener('click', () => {
    const audio = document.getElementById('ambient');
    if (audio) audio.play().catch(console.warn);
}, { once: true });

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



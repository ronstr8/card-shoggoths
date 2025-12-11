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

        if (faceUp && containerId === 'player-hand' && gameState && gameState.game_phase === 'discard') {
            img.onclick = () => toggleDiscard(idx, img);
        }

        container.appendChild(img);
    });
}

function toggleDiscard(idx, img) {
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

    // Auto-fill bet defaults
    if (gameState.game_phase === 'PreDrawBetting' || gameState.game_phase === 'PostDrawBetting') { // Values from GamePhase enum string?
        // Wait, JSON usually serializes int enum as number unless Matcher/Stringer used?
        // GamePhase is int in Go. `json:"game_phase"`
        // I need to check how it serializes. 
        // If I didn't add MarshalJSON, it's an int.
        // Let's assume it is 1, 2, 3... or add string logic.
        // Actually, previous app.js generic strings "bet", "discard".
        // My Go `GamePhase` is a custom type.
        // Let's assume I need to handle integers or check `game.go`.
        // Constants: PhaseAnte=0, PreDrawBetting=1, Discard=2, PostDrawBetting=3, Showdown=4, Complete=5.

        // Actually, just relying on Phase map is safer or standardizing.
        // Let's check `game_phase` values in JS console later or assume ints for now.
        // Standard: 1=Betting1, 2=Discard, 3=Betting2
    }

    // React to phase names from Go (MarshalText returns strings: "ante", "bet_pre", "discard", "bet_post", "showdown", "complete")
    const phase = gameState.game_phase;

    const isBetting = (phase === "bet_pre" || phase === "bet_post" || phase === "bet"); // "bet" legacy support
    const isDiscard = (phase === "discard");
    const isComplete = (phase === "complete" || phase === "end" || phase === "ante" || phase === "deal");

    if (dealBtn) dealBtn.disabled = !isComplete && phase !== "ante" && phase !== "deal";
    if (betBtn) betBtn.disabled = !isBetting;
    if (foldBtn) foldBtn.disabled = (!isBetting && !isDiscard);
    if (discardBtn) discardBtn.disabled = !isDiscard || (gameState.players[0].discarded);
    if (showdownBtn) showdownBtn.disabled = !(phase === "showdown");

    // Input Handling
    if (isBetting) {
        betInput.disabled = false;
        const player = gameState.players[0];
        const toCall = gameState.current_bet - player.bet;

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
        betBtn.textContent = "Bet";
    }
}

async function deal() {
    try {
        const res = await safeFetch('/api/deal');
        gameState = await res.json();

        // Reset local state
        discardIndices = [];

        renderHand('player-hand', gameState.players[0].hand, true);
        renderHand('opponent-hand', gameState.players[1].hand, false); // Hidden
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
    const player = gameState.players[0];
    const toCall = gameState.current_bet - player.bet;

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

        renderHand('player-hand', gameState.players[0].hand, true);
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
        updateSanityDisplay();
        updateButtons();
        document.getElementById('result').textContent = gameState.last_action;

        // If reveal on fold?
        if (gameState.reveal_on_fold) {
            renderHand('opponent-hand', gameState.players[1].hand, true);
        }
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
        renderHand('player-hand', gameState.players[0].hand, true);
        discardIndices = [];
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

        renderHand('opponent-hand', gameState.players[1].hand, true);
        updateSanityDisplay();
        updateButtons();
        document.getElementById('result').textContent = data.result;
    } catch (e) { console.error(e); }
}

// Init
document.addEventListener('DOMContentLoaded', () => {
    updateButtons();
    // Maybe load state on refresh?
    // fetch('/api/state')...
});
window.addEventListener('click', () => {
    const audio = document.getElementById('ambient');
    if (audio) audio.play().catch(console.warn);
}, { once: true });



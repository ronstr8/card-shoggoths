let gameState = null;
let discardIndices = [];
const HTTP_STATUS = {
    UNAUTHORIZED: 401,
    FORBIDDEN: 403
};

// Fetch wrapper with session handling
async function safeFetch(url, opts = {}) {
    opts.credentials = 'same-origin';
    let res = await fetch(url, opts);
    if ((res.status === HTTP_STATUS.UNAUTHORIZED || res.status === HTTP_STATUS.FORBIDDEN) && !opts._retried) {
        await fetch('/api/refresh-session', { method: 'POST', credentials: 'same-origin' });
        opts._retried = true;
        res = await fetch(url, opts);
    }
    return res;
}
// Get sanity emoji based on sanity level
function getSanityEmoji(sanity) {
    if (sanity >= 90) return '😊';
    if (sanity >= 70) return '😐';
    if (sanity >= 50) return '😰';
    if (sanity >= 30) return '😨';
    if (sanity >= 10) return '😱';
    return '🤯';
}
// Update the sanity display
function updateSanityDisplay() {
    if (!gameState) return;

    const playerSanityBar = document.getElementById('player-sanity-bar');
    const playerSanityText = document.getElementById('player-sanity-text');
    const playerSanityEmoji = document.getElementById('player-sanity-emoji');

    const opponentSanityBar = document.getElementById('opponent-sanity-bar');
    const opponentSanityText = document.getElementById('opponent-sanity-text');
    const opponentSanityEmoji = document.getElementById('opponent-sanity-emoji');

    // Update player sanity
    const playerPercent = Math.max(0, (gameState.player_sanity / 100) * 100);
    playerSanityBar.style.width = playerPercent + '%';
    playerSanityText.textContent = gameState.player_sanity;
    playerSanityEmoji.textContent = getSanityEmoji(gameState.player_sanity);

    // Update opponent sanity
    const opponentPercent = Math.max(0, (gameState.opponent_sanity / 100) * 100);
    opponentSanityBar.style.width = opponentPercent + '%';
    opponentSanityText.textContent = gameState.opponent_sanity;
    opponentSanityEmoji.textContent = getSanityEmoji(gameState.opponent_sanity);

    // Update pot
    document.getElementById('pot-amount').textContent = gameState.pot;
}
// Render a hand of cards
function renderHand(containerId, hand, faceUp) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';

    if (!hand || hand.length === 0) return;

    hand.forEach((card, idx) => {
        const img = document.createElement('img');
        img.src = faceUp ? `cards/${card.rank}_of_${card.suit}.png` : 'cards/back.png';
        img.className = 'card';
        img.alt = faceUp ? `${card.rank} of ${card.suit}` : 'Card back';

        if (faceUp && containerId === 'player-hand' && gameState && gameState.game_phase === 'discard') {
            img.onclick = () => toggleDiscard(idx, img);
        }

        container.appendChild(img);
    });
}
// Toggle card selection for discard
function toggleDiscard(idx, img) {
    const i = discardIndices.indexOf(idx);
    if (i >= 0) {
        discardIndices.splice(i, 1);
        img.classList.remove('discard');
    } else {
        if (discardIndices.length >= 3) return; // Max 3 discards
        discardIndices.push(idx);
        img.classList.add('discard');
    }
}
// Update button states based on game phase
function updateButtons() {
    const dealBtn = document.getElementById('deal-btn');
    const betBtn = document.getElementById('bet-btn');
    const foldBtn = document.getElementById('fold-btn');
    const discardBtn = document.getElementById('discard-btn');
    const showdownBtn = document.getElementById('showdown-btn');
    const betInput = document.getElementById('bet-amount');

    if (!gameState) {
        dealBtn.disabled = false;
        betBtn.disabled = true;
        foldBtn.disabled = true;
        discardBtn.disabled = true;
        showdownBtn.disabled = true;
        betInput.disabled = true;
        return;
    }

    // Auto-fill bet amount logic
    if (gameState.game_phase === 'bet') {
        // Calculate amount needed to call/raise
        // Logic: if current_bet > player_bet, we need to call (current - player).
        const toCall = gameState.current_bet - gameState.player_bet;
        let defaultVal = 10;
        if (toCall > 0) {
            defaultVal = toCall;
        }
        // Ensure we update it
        if (betInput) betInput.value = defaultVal.toString();
    } else {
        if (betInput) betInput.value = '10';
    }

    switch (gameState.game_phase) {
        case 'deal':
            dealBtn.disabled = false;
            betBtn.disabled = true;
            foldBtn.disabled = true;
            discardBtn.disabled = true;
            showdownBtn.disabled = true;
            betInput.disabled = true;
            break;
        case 'bet':
            dealBtn.disabled = true;
            betBtn.disabled = false;
            foldBtn.disabled = false;
            discardBtn.disabled = true;
            showdownBtn.disabled = true;
            betInput.disabled = false;
            break;
        case 'discard':
            dealBtn.disabled = true;
            betBtn.disabled = true;
            foldBtn.disabled = false; // Allow folding during discard
            discardBtn.disabled = gameState.discarded;
            showdownBtn.disabled = !gameState.discarded;
            betInput.disabled = true;
            break;
        case 'showdown':
        case 'end':
            dealBtn.disabled = false;
            betBtn.disabled = true;
            foldBtn.disabled = true;
            discardBtn.disabled = true;
            showdownBtn.disabled = true;
            betInput.disabled = true;

            // Reveal opponent cards if configured
            if (gameState && gameState.reveal_on_fold) {
                renderHand('opponent-hand', gameState.opponent_hand, true);
            }
            break;
    }
}
// Deal new cards
async function deal() {
    console.log("Calling API /api/deal");
    try {
        const res = await safeFetch('/api/deal');
        gameState = await res.json();

        discardIndices = [];
        renderHand('player-hand', gameState.player_hand, true);
        renderHand('opponent-hand', gameState.opponent_hand, false);
        updateSanityDisplay();
        updateButtons();

        document.getElementById('result').textContent = '';
        document.getElementById('bet-amount').value = '10'; // Default bet

        // Start betting phase
        gameState.game_phase = 'bet';
        updateButtons();
    } catch (error) {
        console.error('Deal error:', error);
        document.getElementById('result').textContent = 'Error dealing cards';
    }
}
// Fold hand
async function fold() {
    if (!gameState) return;
    console.log("Calling API /api/bet with action: fold");
    try {
        const res = await safeFetch('/api/bet', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: 'fold', amount: 0 })
        });

        const data = await res.json();
        gameState = data.state;

        updateSanityDisplay();
        updateButtons();

        document.getElementById('result').textContent = data.message || 'You folded.';
    } catch (error) {
        console.error('Fold error:', error);
        document.getElementById('result').textContent = 'Error folding';
    }
}

// Place a bet
async function placeBet() {
    const betAmount = parseInt(document.getElementById('bet-amount').value);
    if (!betAmount || betAmount <= 0) {
        document.getElementById('result').textContent = 'Please enter a valid bet amount';
        return;
    }

    if (betAmount > gameState.player_sanity) {
        document.getElementById('result').textContent = 'Not enough sanity to place that bet!';
        return;
    }

    console.log("Calling API /api/bet");
    try {
        // Send explicit amount, action inference handled by backend or we can enhance
        const res = await safeFetch('/api/bet', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ amount: betAmount })
        });

        const data = await res.json();

        if (!res.ok) {
            document.getElementById('result').textContent = data.message || data.error || 'Error placing bet';
            return;
        }

        gameState = data.state;
        updateSanityDisplay();
        // Re-render hand to attach event listeners if we entered discard phase
        renderHand('player-hand', gameState.player_hand, true);
        updateButtons();

        document.getElementById('result').textContent = data.message || 'Bet placed!';

        // If opponent responded, show the result
        if (data.opponent_response) {
            document.getElementById('result').textContent = data.opponent_response;
        }

    } catch (error) {
        console.error('Bet error:', error);
        document.getElementById('result').textContent = 'Error placing bet';
    }
}
// Submit discard selection
async function submitDiscard() {
    if (!gameState || gameState.game_phase !== 'discard') {
        document.getElementById('result').textContent = 'Cannot discard at this time';
        return;
    }

    console.log("Calling API /api/discard");
    try {
        const res = await safeFetch('/api/discard', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ indices: discardIndices })
        });

        gameState = await res.json();
        renderHand('player-hand', gameState.player_hand, true);
        updateButtons();

        document.getElementById('result').textContent = 'Cards discarded!';
        discardIndices = [];

    } catch (error) {
        console.error('Discard error:', error);
        document.getElementById('result').textContent = 'Error discarding cards';
    }
}
// Showdown
async function showdown() {
    if (!gameState || !gameState.discarded) {
        document.getElementById('result').textContent = 'Must discard first';
        return;
    }

    console.log("Calling API /api/showdown");
    try {
        const res = await safeFetch('/api/showdown');
        const data = await res.json();

        gameState = data.state;
        renderHand('opponent-hand', gameState.opponent_hand, true);
        updateSanityDisplay();
        updateButtons();

        document.getElementById('result').textContent = data.result;

    } catch (error) {
        console.error('Showdown error:', error);
        document.getElementById('result').textContent = 'Error in showdown';
    }
}
// Initialize the game on page load
document.addEventListener('DOMContentLoaded', function () {
    updateButtons();

    // Set up bet input validation
    const betInput = document.getElementById('bet-amount');
    betInput.addEventListener('input', function () {
        const value = parseInt(this.value);
        if (gameState && value > gameState.player_sanity) {
            this.value = gameState.player_sanity;
        }
    });
});

// Auto-play audio on first interaction
window.addEventListener('click', () => {
    const audio = document.getElementById('ambient');
    if (audio) audio.play().catch(console.warn);
}, { once: true });


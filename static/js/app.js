
let playerHand = [];
let opponentHand = [];
let discardIndices = [];

function renderHand(containerId, hand, faceUp) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    hand.forEach((card, idx) => {
        const img = document.createElement('img');
        img.src = faceUp ? `cards/${card.rank.toLowerCase()}_of_${card.suit.toLowerCase()}.png` : 'cards/back.png';
        img.className = 'card';
        if (faceUp && containerId === 'player-hand') {
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

function deal() {
    fetch('/api/deal').then(res => res.json()).then(data => {
        playerHand = data.player_hand;
        opponentHand = data.opponent_hand;
        discardIndices = [];
        renderHand('player-hand', playerHand, true);
        renderHand('opponent-hand', opponentHand, false);
        document.getElementById('result').textContent = '';
    });
}

function submitDiscard() {
    fetch('/api/discard', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ indices: discardIndices })
    })
    .then(res => res.json())
    .then(data => {
        playerHand = data.player_hand;
        renderHand('player-hand', playerHand, true);
    });
}

function showdown() {
    fetch('/api/showdown').then(res => res.json()).then(data => {
        renderHand('opponent-hand', data.state.opponent_hand, true);
        document.getElementById('result').textContent = data.result;
    });
}

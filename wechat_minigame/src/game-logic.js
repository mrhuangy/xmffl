const { animalIcons, levelConfigs } = require("./config");

function shuffle(items) {
  const result = items.slice();
  for (let index = result.length - 1; index > 0; index -= 1) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    const value = result[index];
    result[index] = result[swapIndex];
    result[swapIndex] = value;
  }
  return result;
}

function getLevelConfig(levelId) {
  return levelConfigs.find((level) => level.levelId === levelId) || levelConfigs[levelConfigs.length - 1];
}

function createDeck(level) {
  const icons = animalIcons.slice(0, level.pairCount);
  const cards = [];

  icons.forEach((iconKey, pairIndex) => {
    const pairId = `${level.levelId}-${pairIndex}`;
    cards.push({ id: `${pairId}-a`, pairId, iconKey });
    cards.push({ id: `${pairId}-b`, pairId, iconKey });
  });

  return shuffle(cards).map((card, index) => ({
    ...card,
    index,
    state: "face_up"
  }));
}

class MatchGame {
  constructor(levelId) {
    this.level = getLevelConfig(levelId);
    this.cards = createDeck(this.level);
    this.state = "previewing_cards";
    this.openedIndexes = [];
    this.steps = 0;
    this.mismatchCount = 0;
    this.usedHints = 0;
    this.elapsedMs = 0;
    this.startedAt = 0;
    this.lastResult = null;
  }

  finishPreview(now) {
    if (this.state !== "previewing_cards") {
      return;
    }

    this.cards = this.cards.map((card) => ({ ...card, state: "face_down" }));
    this.state = "idle";
    this.startedAt = now;
  }

  tick(now) {
    if (this.startedAt > 0 && this.state !== "previewing_cards" && !this.isEnded()) {
      this.elapsedMs = Math.max(0, now - this.startedAt);
    }

    if (!this.isEnded() && this.elapsedMs > this.level.levelTimeLimitSeconds * 1000) {
      this.state = "level_failed";
      this.lastResult = this.createResult("time_out");
    }
  }

  flipCard(index, now) {
    this.tick(now);

    if (this.state !== "idle" && this.state !== "first_card_opened") {
      return { type: "ignored" };
    }

    const card = this.cards[index];
    if (!card || card.state !== "face_down") {
      return { type: "ignored" };
    }

    this.cards[index] = { ...card, state: "face_up" };
    this.openedIndexes.push(index);

    if (this.openedIndexes.length === 1) {
      this.state = "first_card_opened";
      return { type: "first_card_opened" };
    }

    this.steps += 1;
    const [firstIndex, secondIndex] = this.openedIndexes;
    const first = this.cards[firstIndex];
    const second = this.cards[secondIndex];

    if (first.pairId === second.pairId) {
      this.cards[firstIndex] = { ...first, state: "removed" };
      this.cards[secondIndex] = { ...second, state: "removed" };
      this.openedIndexes = [];

      if (this.cards.every((item) => item.state === "removed")) {
        this.state = "level_complete";
        this.lastResult = this.createResult("completed");
        return { type: "level_complete", result: this.lastResult };
      }

      this.state = "idle";
      return {
        type: "matched",
        cards: [
          { index: firstIndex, iconKey: first.iconKey },
          { index: secondIndex, iconKey: second.iconKey }
        ]
      };
    }

    this.mismatchCount += 1;
    this.state = "resolving_mismatch";

    if (this.mismatchCount >= this.level.maxMismatchCount) {
      this.state = "level_failed";
      this.lastResult = this.createResult("mismatch_limit");
      return { type: "level_failed", result: this.lastResult };
    }

    return { type: "mismatched" };
  }

  finishMismatch() {
    if (this.state !== "resolving_mismatch") {
      return;
    }

    for (const index of this.openedIndexes) {
      const card = this.cards[index];
      if (card && card.state === "face_up") {
        this.cards[index] = { ...card, state: "face_down" };
      }
    }

    this.openedIndexes = [];
    this.state = "idle";
  }

  useHint() {
    if (this.state !== "idle" && this.state !== "first_card_opened") {
      return [];
    }

    const candidates = this.cards.filter((card) => card.state === "face_down");
    const first = candidates.find((card) =>
      candidates.some((other) => other.id !== card.id && other.pairId === card.pairId)
    );
    const second = first && candidates.find((card) => card.id !== first.id && card.pairId === first.pairId);

    if (!first || !second) {
      return [];
    }

    this.usedHints += 1;
    return [first.index, second.index];
  }

  revealUnmatched() {
    if (this.state !== "idle") {
      return [];
    }

    const indexes = [];
    this.cards = this.cards.map((card, index) => {
      if (card.state !== "face_down") {
        return card;
      }

      indexes.push(index);
      return { ...card, state: "face_up" };
    });
    return indexes;
  }

  hideRevealed(indexes) {
    const targetIndexes = indexes || [];
    this.cards = this.cards.map((card, index) => {
      if (targetIndexes.includes(index) && card.state === "face_up") {
        return { ...card, state: "face_down" };
      }

      return card;
    });
  }

  removeOnePair() {
    if (this.state !== "idle") {
      return { type: "ignored" };
    }

    const candidates = this.cards.filter((card) => card.state === "face_down");
    const first = candidates.find((card) =>
      candidates.some((other) => other.id !== card.id && other.pairId === card.pairId)
    );
    const second = first && candidates.find((card) => card.id !== first.id && card.pairId === first.pairId);

    if (!first || !second) {
      return { type: "ignored" };
    }

    this.cards[first.index] = { ...this.cards[first.index], state: "removed" };
    this.cards[second.index] = { ...this.cards[second.index], state: "removed" };

    if (this.cards.every((item) => item.state === "removed")) {
      this.state = "level_complete";
      this.lastResult = this.createResult("completed");
      return { type: "level_complete", result: this.lastResult };
    }

    return { type: "removed_pair" };
  }

  createResult(reason) {
    const success = reason === "completed";
    const stars = success ? this.calculateStars() : 0;
    return {
      levelId: this.level.levelId,
      success,
      reason,
      steps: this.steps,
      mismatchCount: this.mismatchCount,
      elapsedMs: this.elapsedMs,
      stars,
      coinsEarned: success ? stars * 10 : 0,
      usedHints: this.usedHints,
      completedAt: Date.now()
    };
  }

  calculateStars() {
    const elapsedSeconds = Math.ceil(this.elapsedMs / 1000);
    if (this.steps <= this.level.excellentStepThreshold && elapsedSeconds <= this.level.excellentTimeThreshold) {
      return 3;
    }
    if (this.steps <= this.level.normalStepThreshold && elapsedSeconds <= this.level.normalTimeThreshold) {
      return 2;
    }
    return 1;
  }

  remainingSeconds() {
    return Math.max(0, this.level.levelTimeLimitSeconds - Math.floor(this.elapsedMs / 1000));
  }

  isEnded() {
    return this.state === "level_complete" || this.state === "level_failed";
  }
}

module.exports = {
  MatchGame,
  getLevelConfig,
  levelConfigs
};

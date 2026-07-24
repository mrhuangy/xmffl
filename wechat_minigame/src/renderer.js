function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}

function roundRect(ctx, x, y, width, height, radius) {
  const r = Math.min(radius, width / 2, height / 2);
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + width, y, x + width, y + height, r);
  ctx.arcTo(x + width, y + height, x, y + height, r);
  ctx.arcTo(x, y + height, x, y, r);
  ctx.arcTo(x, y, x + width, y, r);
  ctx.closePath();
}

function drawImageCover(ctx, image, x, y, width, height) {
  if (!image) {
    ctx.fillStyle = "#e6f6ff";
    ctx.fillRect(x, y, width, height);
    return;
  }

  const imageRatio = image.width / image.height;
  const boxRatio = width / height;
  let sourceWidth = image.width;
  let sourceHeight = image.height;
  let sourceX = 0;
  let sourceY = 0;

  if (imageRatio > boxRatio) {
    sourceWidth = image.height * boxRatio;
    sourceX = (image.width - sourceWidth) / 2;
  } else {
    sourceHeight = image.width / boxRatio;
    sourceY = (image.height - sourceHeight) / 2;
  }

  ctx.drawImage(image, sourceX, sourceY, sourceWidth, sourceHeight, x, y, width, height);
}

function drawImageContain(ctx, image, x, y, width, height) {
  if (!image) {
    return;
  }

  const rect = containRect(image, x, y, width, height);
  ctx.drawImage(image, rect.x, rect.y, rect.w, rect.h);
}

function containRect(image, x, y, width, height) {
  if (!image) {
    return { x, y, w: width, h: height, scale: 1, frameX: x, frameY: y, frameW: width, frameH: height };
  }

  const scale = Math.min(width / image.width, height / image.height);
  const drawWidth = image.width * scale;
  const drawHeight = image.height * scale;
  return {
    x: x + (width - drawWidth) / 2,
    y: y + (height - drawHeight) / 2,
    w: drawWidth,
    h: drawHeight,
    scale,
    frameX: x,
    frameY: y,
    frameW: width,
    frameH: height
  };
}

function fillTextCenter(ctx, text, x, y, width, size, color, weight = "700") {
  ctx.fillStyle = color;
  ctx.font = `${weight} ${size}px sans-serif`;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(text, x + width / 2, y);
}

function strokeTextCenter(ctx, text, x, y, width, size, color, strokeColor, lineWidth = 4, weight = "900") {
  ctx.font = `${weight} ${size}px sans-serif`;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.lineJoin = "round";
  ctx.strokeStyle = strokeColor;
  ctx.lineWidth = lineWidth;
  ctx.strokeText(text, x + width / 2, y);
  ctx.fillStyle = color;
  ctx.fillText(text, x + width / 2, y);
}

function ellipsize(text, maxLength) {
  const chars = Array.from(String(text || ""));
  if (chars.length <= maxLength) {
    return chars.join("");
  }
  return `${chars.slice(0, Math.max(0, maxLength - 1)).join("")}...`;
}

class Renderer {
  constructor(ctx, assets) {
    this.ctx = ctx;
    this.assets = assets;
  }

  clear(width, height) {
    this.ctx.clearRect(0, 0, width, height);
  }

  home(state) {
    const {
      width,
      height,
      buttons,
      progress,
      player,
      toast,
      settingsOpen,
      coinAdDialogOpen,
      staminaShopOpen,
      staminaExchangeConfirm,
      settings,
      unlimitedStamina
    } = state;
    const ctx = this.ctx;
    drawImageCover(ctx, this.assets.image("homeBg"), 0, 0, width, height);

    this.homeHud(width, height, buttons, progress, player, unlimitedStamina);

    drawImageContain(ctx, this.assets.image("logo"), width * 0.065, height * 0.145, width * 0.88, height * 0.27);

    drawImageContain(ctx, this.assets.image("startButton"), buttons.start.x, buttons.start.y, buttons.start.w, buttons.start.h);
    drawImageContain(ctx, this.assets.image("selectButton"), buttons.levels.x, buttons.levels.y, buttons.levels.w, buttons.levels.h);
    drawImageContain(ctx, this.assets.image("timeModeButton"), buttons.timeMode.x, buttons.timeMode.y, buttons.timeMode.w, buttons.timeMode.h);
    strokeTextCenter(
      ctx,
      "\u5f00\u59cb\u6e38\u620f",
      buttons.start.x,
      buttons.start.y + buttons.start.h * 0.72,
      buttons.start.w,
      Math.max(20, Math.min(30, buttons.start.w * 0.12)),
      "#ffffff",
      "#b96a09",
      5
    );
    strokeTextCenter(
      ctx,
      `\u7b2c ${progress.currentLevel} \u5173`,
      buttons.start.x,
      buttons.start.y + buttons.start.h * 0.85,
      buttons.start.w,
      Math.max(13, Math.min(18, buttons.start.w * 0.072)),
      "#ffffff",
      "#c47a12",
      3,
      "800"
    );
    strokeTextCenter(
      ctx,
      "\u9009\u62e9\u5173\u5361",
      buttons.levels.x,
      buttons.levels.y + buttons.levels.h * 0.5,
      buttons.levels.w,
      Math.max(20, Math.min(28, buttons.levels.w * 0.11)),
      "#ffffff",
      "#3f8b19",
      5
    );
    strokeTextCenter(
      ctx,
      "\u65e0\u5c3d\u6a21\u5f0f",
      buttons.timeMode.x,
      buttons.timeMode.y + buttons.timeMode.h * 0.5,
      buttons.timeMode.w,
      Math.max(20, Math.min(28, buttons.timeMode.w * 0.11)),
      "#ffffff",
      "#2e77a8",
      5
    );

    if (toast) {
      this.customToast(width, height, toast.text);
    }

    if (settingsOpen) {
      this.settingsPanel(width, height, buttons, settings);
    }

    if (coinAdDialogOpen) {
      this.coinAdDialog(width, height, buttons);
    }

    if (!unlimitedStamina && staminaShopOpen) {
      this.staminaShop(width, height, buttons);
    }

    if (!unlimitedStamina && staminaExchangeConfirm) {
      this.staminaExchangeDialog(width, height, buttons, staminaExchangeConfirm);
    }
  }

  customToast(width, height, text) {
    const ctx = this.ctx;
    const boxW = Math.min(width * 0.72, 280);
    const boxH = 54;
    const x = (width - boxW) / 2;
    const y = height * 0.32;
    ctx.save();
    ctx.fillStyle = "rgba(40, 62, 43, 0.88)";
    roundRect(ctx, x, y, boxW, boxH, 14);
    ctx.fill();
    ctx.strokeStyle = "rgba(255,255,255,0.42)";
    ctx.lineWidth = 2;
    ctx.stroke();
    fillTextCenter(ctx, text, x, y + boxH / 2, boxW, 17, "#ffffff", "800");
    ctx.restore();
  }

  settingsPanel(width, height, buttons, settings) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.36)";
    ctx.fillRect(0, 0, width, height);

    const panelW = Math.min(width * 0.78, 320);
    const panelH = 230;
    const x = (width - panelW) / 2;
    const y = height * 0.28;
    ctx.fillStyle = "#fff9df";
    roundRect(ctx, x, y, panelW, panelH, 18);
    ctx.fill();
    ctx.strokeStyle = "#c8842c";
    ctx.lineWidth = 4;
    ctx.stroke();

    strokeTextCenter(ctx, "\u64cd\u4f5c", x, y + 38, panelW, 24, "#ffffff", "#8f4f17", 4, "900");
    this.settingsSwitch(buttons.settingsMusic, "\u97f3\u4e50", settings.musicEnabled);
    this.settingsSwitch(buttons.settingsSfx, "\u97f3\u6548", settings.sfxEnabled);
    this.settingsCloseButton(buttons.settingsClose);
    ctx.restore();
  }

  settingsSwitch(rect, label, enabled) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "#ffffff";
    roundRect(ctx, rect.x, rect.y, rect.w, rect.h, 12);
    ctx.fill();
    ctx.strokeStyle = "#e2b15a";
    ctx.lineWidth = 2;
    ctx.stroke();

    ctx.fillStyle = "#7b3f17";
    ctx.font = "900 18px sans-serif";
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillText(label, rect.x + 18, rect.y + rect.h / 2);

    const switchW = 62;
    const switchH = 28;
    const sx = rect.x + rect.w - switchW - 14;
    const sy = rect.y + (rect.h - switchH) / 2;
    ctx.fillStyle = enabled ? "#78d64b" : "#b8b8b8";
    roundRect(ctx, sx, sy, switchW, switchH, switchH / 2);
    ctx.fill();
    ctx.strokeStyle = enabled ? "#4ba733" : "#8b8b8b";
    ctx.lineWidth = 2;
    ctx.stroke();

    const knobR = switchH * 0.39;
    const knobX = enabled ? sx + switchW - switchH / 2 : sx + switchH / 2;
    ctx.fillStyle = "#ffffff";
    ctx.beginPath();
    ctx.arc(knobX, sy + switchH / 2, knobR, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#ffffff";
    ctx.font = "800 11px sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(enabled ? "ON" : "OFF", enabled ? sx + switchW * 0.28 : sx + switchW * 0.72, sy + switchH / 2);
    ctx.restore();
  }

  settingsCloseButton(rect) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "#f5b43c";
    roundRect(ctx, rect.x, rect.y, rect.w, rect.h, 14);
    ctx.fill();
    ctx.strokeStyle = "#b96a09";
    ctx.lineWidth = 3;
    ctx.stroke();
    strokeTextCenter(ctx, "\u5173\u95ed", rect.x, rect.y + rect.h / 2, rect.w, 18, "#ffffff", "#9d5a12", 3, "900");
    ctx.restore();
  }

  coinAdDialog(width, height, buttons) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.36)";
    ctx.fillRect(0, 0, width, height);

    const panelW = Math.min(width * 0.78, 320);
    const panelH = 188;
    const x = (width - panelW) / 2;
    const y = height * 0.31;
    ctx.fillStyle = "#fff9df";
    roundRect(ctx, x, y, panelW, panelH, 18);
    ctx.fill();
    ctx.strokeStyle = "#c8842c";
    ctx.lineWidth = 4;
    ctx.stroke();

    strokeTextCenter(ctx, "\u83b7\u53d6\u91d1\u5e01", x, y + 36, panelW, 23, "#ffffff", "#8f4f17", 4, "900");
    fillTextCenter(ctx, "\u89c2\u770b\u5e7f\u544a\u53ef\u83b7\u53d6100\u91d1\u5e01", x, y + 86, panelW, 16, "#7b3f17", "800");

    this.dialogButton(buttons.coinAdCancel, "\u53d6\u6d88", "#d5d0c4", "#8a8275");
    this.dialogButton(buttons.coinAdConfirm, "\u786e\u5b9a", "#f5b43c", "#b96a09");
    ctx.restore();
  }

  staminaShop(width, height, buttons) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.36)";
    ctx.fillRect(0, 0, width, height);

    const panelW = Math.min(width * 0.76, 320);
    const panelH = 294;
    const x = (width - panelW) / 2;
    const y = buttons.staminaAd.y - 58;
    ctx.fillStyle = "#fff9df";
    roundRect(ctx, x, y, panelW, panelH, 18);
    ctx.fill();
    ctx.strokeStyle = "#c8842c";
    ctx.lineWidth = 4;
    ctx.stroke();

    strokeTextCenter(ctx, "\u8d2d\u4e70\u4f53\u529b", x, y + 36, panelW, 23, "#ffffff", "#8f4f17", 4, "900");
    this.shopRow(buttons.staminaAd, "\u89c2\u770b\u5e7f\u544a", "\u83b7\u5f973\u4f53\u529b", "#78d64b", "#4ba733");
    this.shopRow(buttons.staminaBuy1, "99\u91d1\u5e01", "\u83b7\u5f971\u4f53\u529b", "#f5b43c", "#b96a09");
    this.shopRow(buttons.staminaBuy3, "266\u91d1\u5e01", "\u83b7\u5f973\u4f53\u529b", "#f5b43c", "#b96a09");
    this.shopRow(buttons.staminaBuy5, "388\u91d1\u5e01", "\u83b7\u5f975\u4f53\u529b", "#f5b43c", "#b96a09");
    this.dialogButton(buttons.staminaShopClose, "\u5173\u95ed", "#d5d0c4", "#8a8275");
    ctx.restore();
  }

  shopRow(rect, title, desc, fill, stroke) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = fill;
    roundRect(ctx, rect.x, rect.y, rect.w, rect.h, 13);
    ctx.fill();
    ctx.strokeStyle = stroke;
    ctx.lineWidth = 3;
    ctx.stroke();

    ctx.fillStyle = "#ffffff";
    ctx.font = "900 16px sans-serif";
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.lineJoin = "round";
    ctx.strokeStyle = stroke;
    ctx.lineWidth = 3;
    ctx.strokeText(title, rect.x + 18, rect.y + rect.h / 2);
    ctx.fillText(title, rect.x + 18, rect.y + rect.h / 2);

    ctx.font = "800 14px sans-serif";
    ctx.textAlign = "right";
    ctx.strokeText(desc, rect.x + rect.w - 18, rect.y + rect.h / 2);
    ctx.fillText(desc, rect.x + rect.w - 18, rect.y + rect.h / 2);
    ctx.restore();
  }

  staminaExchangeDialog(width, height, buttons, exchange) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.36)";
    ctx.fillRect(0, 0, width, height);

    const panelW = Math.min(width * 0.78, 320);
    const panelH = 188;
    const x = (width - panelW) / 2;
    const y = height * 0.31;
    ctx.fillStyle = "#fff9df";
    roundRect(ctx, x, y, panelW, panelH, 18);
    ctx.fill();
    ctx.strokeStyle = "#c8842c";
    ctx.lineWidth = 4;
    ctx.stroke();

    strokeTextCenter(ctx, "\u786e\u8ba4\u5151\u6362", x, y + 36, panelW, 23, "#ffffff", "#8f4f17", 4, "900");
    fillTextCenter(
      ctx,
      `\u6d88\u8017${exchange.cost}\u91d1\u5e01\u5151\u6362${exchange.amount}\u4f53\u529b\uff1f`,
      x,
      y + 86,
      panelW,
      16,
      "#7b3f17",
      "800"
    );

    this.dialogButton(buttons.staminaExchangeCancel, "\u53d6\u6d88", "#d5d0c4", "#8a8275");
    this.dialogButton(buttons.staminaExchangeConfirm, "\u786e\u5b9a", "#f5b43c", "#b96a09");
    ctx.restore();
  }

  dialogButton(rect, text, fill, stroke) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = fill;
    roundRect(ctx, rect.x, rect.y, rect.w, rect.h, 13);
    ctx.fill();
    ctx.strokeStyle = stroke;
    ctx.lineWidth = 3;
    ctx.stroke();
    strokeTextCenter(ctx, text, rect.x, rect.y + rect.h / 2, rect.w, 17, "#ffffff", stroke, 3, "900");
    ctx.restore();
  }

  homeHud(width, height, buttons, progress, player, unlimitedStamina) {
    const ctx = this.ctx;
    const top = Math.max(34, height * 0.05);
    const playerH = Math.max(38, Math.min(46, width * 0.11));
    const playerW = Math.max(94, Math.min(116, width * 0.29));
    const avatar = playerH * 1.06;
    const playerX = 14;
    const gap = Math.max(18, Math.min(22, width * 0.052));
    const statH = Math.max(26, Math.min(30, playerH * 0.68));
    const staminaW = Math.max(62, Math.min(70, width * 0.165));
    const staminaX = playerX + playerW + Math.max(8, width * 0.02);
    const statY = top + Math.max(2, height * 0.008);
    const coinX = unlimitedStamina ? staminaX : staminaX + staminaW + gap;
    const safeRight = width - 92;
    const coinW = Math.max(62, Math.min(Math.max(66, Math.min(74, width * 0.175)), safeRight - coinX));

    this.playerPlate(playerX, top, playerW, playerH, avatar, player);
    if (!unlimitedStamina) {
      const stamina = progress.stamina ?? 5;
      this.resourcePlate(staminaX, statY, staminaW, statH, "heart", String(stamina), this.staminaStatusText(progress));
    }
    this.resourcePlate(coinX, statY, coinW, statH, "coin", String(progress.coins || 0), "");
    this.operationButton(buttons.operation);
  }

  playerPlate(x, y, w, h, avatarSize, player) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(95,54,19,0.25)";
    roundRect(ctx, x + 2, y + 3, w, h, 14);
    ctx.fill();
    const bodyX = x + avatarSize * 0.46;
    const bodyW = w - avatarSize * 0.46;
    const gradient = ctx.createLinearGradient(bodyX, y, bodyX, y + h);
    gradient.addColorStop(0, "#d69742");
    gradient.addColorStop(1, "#9f6128");
    ctx.fillStyle = gradient;
    roundRect(ctx, bodyX, y + 2, bodyW, h - 4, 15);
    ctx.fill();
    ctx.strokeStyle = "#fff0bf";
    ctx.lineWidth = 2.5;
    ctx.stroke();

    ctx.fillStyle = "#fff8d6";
    roundRect(ctx, x, y, avatarSize, avatarSize, 14);
    ctx.fill();
    ctx.strokeStyle = "#b9762c";
    ctx.lineWidth = 3;
    ctx.stroke();
    ctx.save();
    roundRect(ctx, x + 5, y + 5, avatarSize - 10, avatarSize - 10, 12);
    ctx.clip();
    drawImageCover(ctx, this.assets.image("animal:panda"), x + 5, y + 5, avatarSize - 10, avatarSize - 10);
    ctx.restore();

    const nickname = ellipsize((player && player.nickname) || "\u73a9\u5bb6", 5);
    strokeTextCenter(ctx, nickname, bodyX + 14, y + h * 0.36, bodyW - 22, Math.max(15, h * 0.31), "#ffffff", "#8f4f17", 3, "900");
    this.starBadge(bodyX + 18, y + h * 0.72, h * 0.22);
    ctx.fillStyle = "#93de56";
    roundRect(ctx, bodyX + 38, y + h * 0.65, bodyW - 58, h * 0.16, h * 0.08);
    ctx.fill();
    ctx.strokeStyle = "#6aa532";
    ctx.lineWidth = 1.5;
    ctx.stroke();
    ctx.restore();
  }

  resourcePlate(x, y, w, h, type, value, subText) {
    const ctx = this.ctx;
    const iconSize = h * 0.72;
    const plusRadius = h * 0.25;
    const valueX = x + iconSize * 1.18;
    const valueMaxW = Math.max(24, w - iconSize - plusRadius * 2.25);
    ctx.save();
    ctx.fillStyle = "rgba(86,58,24,0.18)";
    roundRect(ctx, x + iconSize * 0.52, y + 3, w - iconSize * 0.42, h, h / 2);
    ctx.fill();
    ctx.fillStyle = "#fff6d9";
    roundRect(ctx, x + iconSize * 0.48, y, w - iconSize * 0.48, h, h / 2);
    ctx.fill();
    ctx.strokeStyle = "#f2d999";
    ctx.lineWidth = 2;
    ctx.stroke();

    if (type === "heart") {
      this.heartIcon(x + h * 0.03, y + h * 0.08, iconSize);
    } else {
      this.bigCoinIcon(x + h * 0.03, y + h * 0.07, iconSize);
    }

    ctx.fillStyle = "#7b3f17";
    ctx.font = `900 ${Math.max(13, h * 0.5)}px sans-serif`;
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillText(value, valueX, y + h * 0.43, valueMaxW);
    if (subText) {
      ctx.font = `800 ${Math.max(6, h * 0.22)}px sans-serif`;
      ctx.fillText(subText, valueX, y + h * 0.78, valueMaxW);
    }
    this.plusButton(x + w - plusRadius * 0.9, y + h * 0.5, plusRadius);
    ctx.restore();
  }

  operationButton(rect) {
    const ctx = this.ctx;
    const cx = rect.x + rect.w / 2;
    const cy = rect.y + rect.h / 2;
    const r = rect.w / 2;
    ctx.save();
    ctx.fillStyle = "rgba(100,60,18,0.24)";
    ctx.beginPath();
    ctx.arc(cx + 2, cy + 3, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = "#b46a26";
    ctx.beginPath();
    ctx.arc(cx, cy, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#fff2c6";
    ctx.lineWidth = 3;
    ctx.stroke();
    ctx.translate(cx, cy);
    ctx.strokeStyle = "#ffffff";
    ctx.lineWidth = Math.max(3, r * 0.14);
    for (let i = 0; i < 8; i += 1) {
      ctx.rotate(Math.PI / 4);
      ctx.beginPath();
      ctx.moveTo(0, -r * 0.25);
      ctx.lineTo(0, -r * 0.52);
      ctx.stroke();
    }
    ctx.beginPath();
    ctx.arc(0, 0, r * 0.25, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();
  }

  heartIcon(x, y, size) {
    const ctx = this.ctx;
    const cx = x + size / 2;
    const cy = y + size * 0.52;
    ctx.save();
    ctx.fillStyle = "#fb614c";
    ctx.strokeStyle = "#e33b30";
    ctx.lineWidth = Math.max(1.3, size * 0.045);
    ctx.beginPath();
    ctx.moveTo(cx, cy + size * 0.28);
    ctx.bezierCurveTo(cx - size * 0.55, cy - size * 0.08, cx - size * 0.36, cy - size * 0.48, cx, cy - size * 0.25);
    ctx.bezierCurveTo(cx + size * 0.36, cy - size * 0.48, cx + size * 0.55, cy - size * 0.08, cx, cy + size * 0.28);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = "rgba(255,255,255,0.45)";
    ctx.beginPath();
    ctx.arc(cx - size * 0.16, cy - size * 0.16, size * 0.1, 0, Math.PI * 2);
    ctx.fill();
    ctx.restore();
  }

  bigCoinIcon(x, y, size) {
    const ctx = this.ctx;
    const cx = x + size / 2;
    const cy = y + size / 2;
    const r = size * 0.42;
    ctx.save();
    ctx.fillStyle = "#ffc832";
    ctx.beginPath();
    ctx.arc(cx, cy, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#e7981e";
    ctx.lineWidth = Math.max(1.5, size * 0.06);
    ctx.stroke();
    ctx.fillStyle = "#ffdf62";
    ctx.beginPath();
    ctx.arc(cx, cy, r * 0.68, 0, Math.PI * 2);
    ctx.fill();
    this.starPath(cx, cy, r * 0.38, "#f3b42a");
    ctx.restore();
  }

  plusButton(cx, cy, r) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "#88d853";
    ctx.beginPath();
    ctx.arc(cx, cy, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#4ba733";
    ctx.lineWidth = Math.max(1.4, r * 0.16);
    ctx.stroke();
    ctx.strokeStyle = "#ffffff";
    ctx.lineWidth = Math.max(2, r * 0.22);
    ctx.lineCap = "round";
    ctx.beginPath();
    ctx.moveTo(cx - r * 0.42, cy);
    ctx.lineTo(cx + r * 0.42, cy);
    ctx.moveTo(cx, cy - r * 0.42);
    ctx.lineTo(cx, cy + r * 0.42);
    ctx.stroke();
    ctx.restore();
  }

  starBadge(cx, cy, r) {
    this.starPath(cx, cy, r, "#ffd84d", "#bd7620");
  }

  starPath(cx, cy, r, fill, stroke) {
    const ctx = this.ctx;
    ctx.save();
    ctx.beginPath();
    for (let i = 0; i < 10; i += 1) {
      const angle = -Math.PI / 2 + (i * Math.PI) / 5;
      const radius = i % 2 === 0 ? r : r * 0.48;
      const x = cx + Math.cos(angle) * radius;
      const y = cy + Math.sin(angle) * radius;
      if (i === 0) {
        ctx.moveTo(x, y);
      } else {
        ctx.lineTo(x, y);
      }
    }
    ctx.closePath();
    ctx.fillStyle = fill;
    ctx.fill();
    if (stroke) {
      ctx.strokeStyle = stroke;
      ctx.lineWidth = 1.5;
      ctx.stroke();
    }
    ctx.restore();
  }

  loading(state) {
    const { width, height, progress } = state;
    const ctx = this.ctx;
    ctx.fillStyle = "#3f8f4f";
    ctx.fillRect(0, 0, width, height);
    drawImageCover(ctx, this.assets.image("loadingBg"), 0, 0, width, height);
    drawImageContain(ctx, this.assets.image("logo"), width * 0.065, height * 0.09, width * 0.88, height * 0.28);

    const loadingImage = this.assets.image("loading");
    const box = containRect(loadingImage, width * 0.13, height * 0.42, width * 0.74, height * 0.28);
    drawImageContain(ctx, loadingImage, box.frameX, box.frameY, box.frameW, box.frameH);

    if (!loadingImage) {
      this.drawButton({ x: width * 0.24, y: height * 0.55, w: width * 0.52, h: 30 }, "Loading", "#8bd94a");
      return;
    }

    const barX = box.x + box.scale * 148;
    const barY = box.y + box.scale * 250;
    const barW = box.scale * 316;
    const barH = box.scale * 22;
    const inset = Math.max(2, box.scale * 2);
    const fillW = Math.max(barH, (barW - inset * 2) * Math.max(0.03, Math.min(1, progress)));

    ctx.fillStyle = "#77d83f";
    roundRect(ctx, barX + inset, barY + inset, fillW, barH - inset * 2, (barH - inset * 2) / 2);
    ctx.fill();
  }

  levels(state) {
    const {
      width,
      height,
      buttons,
      levels,
      currentPage,
      pageCount,
      progress,
      toast,
      staminaShopOpen,
      staminaExchangeConfirm,
      unlimitedStamina
    } = state;
    const ctx = this.ctx;
    drawImageCover(ctx, this.assets.image("homeBg"), 0, 0, width, height);
    drawImageContain(ctx, this.assets.image("back"), buttons.back.x, buttons.back.y, buttons.back.w, buttons.back.h);
    if (!unlimitedStamina) {
      this.levelTopStamina(width, height, progress);
    }

    const titleW = Math.min(width * 0.76, 380);
    const titleH = titleW * (576 / 1824);
    drawImageContain(ctx, this.assets.image("selectTitle"), (width - titleW) / 2, height * 0.13, titleW, titleH);

    const panelX = width * 0.075;
    const panelY = height * 0.255;
    const panelW = width * 0.85;
    const panelH = Math.min(height * 0.56, height - panelY - 118);
    this.levelPanel(panelX, panelY, panelW, panelH);
    this.levelTabs(panelX, panelY, panelW);

    for (const card of levels) {
      this.levelCard(card);
    }

    this.levelPager(width, panelY + panelH - 43, currentPage, pageCount);
    if (!unlimitedStamina && staminaShopOpen) {
      this.staminaShop(width, height, buttons);
    }

    if (!unlimitedStamina && staminaExchangeConfirm) {
      this.staminaExchangeDialog(width, height, buttons, staminaExchangeConfirm);
    }

    if (toast) {
      this.customToast(width, height, toast.text);
    }
  }

  levelTopStamina(width, height, progress) {
    const statH = Math.max(26, Math.min(30, width * 0.075));
    const staminaW = Math.max(62, Math.min(70, width * 0.165));
    const staminaX = width - staminaW - 112;
    const staminaY = Math.max(44, height * 0.055);
    const stamina = progress.stamina ?? 5;
    const maxStamina = progress.maxStamina ?? 5;
    this.resourcePlate(staminaX, staminaY, staminaW, statH, "heart", String(stamina), this.staminaStatusText(progress));
  }

  staminaStatusText(progress) {
    const stamina = progress.stamina ?? 5;
    const maxStamina = progress.maxStamina ?? 5;
    if (stamina >= maxStamina) {
      return "\u5df2\u6ee1";
    }
    const nextRecoverAt = progress.nextStaminaRecoverAt || 0;
    const remainMs = nextRecoverAt - Date.now();
    if (remainMs <= 0) {
      return "00:00";
    }
    const totalSeconds = Math.ceil(remainMs / 1000);
    const minute = Math.floor(totalSeconds / 60);
    const second = totalSeconds % 60;
    return `${String(minute).padStart(2, "0")}:${String(second).padStart(2, "0")}`;
  }

  levelPanel(x, y, w, h) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(108, 65, 22, 0.18)";
    roundRect(ctx, x + 2, y + 5, w, h, 25);
    ctx.fill();
    ctx.fillStyle = "#fff6dc";
    roundRect(ctx, x, y, w, h, 25);
    ctx.fill();
    ctx.strokeStyle = "#edc982";
    ctx.lineWidth = 4;
    ctx.stroke();
    ctx.fillStyle = "rgba(255, 255, 255, 0.35)";
    roundRect(ctx, x + 8, y + 8, w - 16, h - 16, 19);
    ctx.fill();
    ctx.restore();
  }

  levelTabs(x, y, w) {
    const ctx = this.ctx;
    const tabY = y + 22;
    const tabW = w * 0.42;
    const tabH = 44;
    const gap = w * 0.035;
    ctx.save();
    ctx.fillStyle = "#9fde4c";
    roundRect(ctx, x + w * 0.07, tabY, tabW, tabH, 15);
    ctx.fill();
    ctx.strokeStyle = "#79bd32";
    ctx.lineWidth = 3;
    ctx.stroke();
    strokeTextCenter(ctx, "\u666e\u901a\u5173\u5361", x + w * 0.07, tabY + tabH / 2, tabW, 20, "#ffffff", "#5f9e2a", 3, "900");

    ctx.fillStyle = "#f5dfba";
    roundRect(ctx, x + w * 0.07 + tabW + gap, tabY, tabW, tabH, 15);
    ctx.fill();
    ctx.strokeStyle = "#e6bf84";
    ctx.lineWidth = 2;
    ctx.stroke();
    fillTextCenter(ctx, "\u6311\u6218\u5173\u5361", x + w * 0.07 + tabW + gap, tabY + tabH / 2, tabW, 18, "#8a4f20", "900");
    drawImageContain(ctx, this.assets.image("lock"), x + w * 0.07 + tabW + gap + tabW * 0.82, tabY + 11, 22, 22);
    ctx.restore();
  }

  levelCard(card) {
    const ctx = this.ctx;
    const rect = card.rect;
    const image = card.isCurrent
      ? this.assets.image("blockNow")
      : this.assets.image(card.unlocked ? "blockAlready" : "blockNo");

    ctx.save();
    drawImageContain(ctx, image, rect.x, rect.y, rect.w, rect.h);
    strokeTextCenter(
      ctx,
      String(card.level.levelId),
      rect.x,
      rect.y + rect.h * 0.33,
      rect.w,
      Math.max(21, rect.w * 0.43),
      "#ffffff",
      "#7d4b20",
      3,
      "900"
    );

    if (card.unlocked) {
      const starSize = rect.w * 0.25;
      const starGap = rect.w * 0.24;
      const startX = rect.x + rect.w / 2 - starGap;
      for (let index = 0; index < 3; index += 1) {
        const star = index < card.stars ? this.assets.image("winStar") : this.assets.image("lossStar");
        drawImageContain(ctx, star, startX + index * starGap - starSize / 2, rect.y + rect.h * 0.65, starSize, starSize);
      }
    } else {
      const lockSize = rect.w * 0.42;
      drawImageContain(ctx, this.assets.image("lock"), rect.x + rect.w / 2 - lockSize / 2, rect.y + rect.h * 0.5, lockSize, lockSize);
    }
    ctx.restore();
  }

  levelPager(width, y, currentPage, pageCount) {
    const ctx = this.ctx;
    ctx.save();
    this.pagerArrow(width * 0.27, y, 1);
    this.pagerArrow(width * 0.73, y, -1);
    const dotY = y + 7;
    const count = Math.max(1, pageCount || 1);
    const gap = 11;
    const startX = width / 2 - ((count - 1) * gap) / 2;
    for (let index = 0; index < count; index += 1) {
      ctx.fillStyle = index === currentPage ? "#92d64a" : "#e4c99d";
      ctx.beginPath();
      ctx.arc(startX + index * gap, dotY, 4.5, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.restore();
  }

  pagerArrow(cx, y, dir) {
    const ctx = this.ctx;
    ctx.save();
    ctx.translate(cx, y + 7);
    ctx.scale(dir, 1);
    ctx.fillStyle = "#fff1d3";
    ctx.strokeStyle = "#b7762e";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(-10, 0);
    ctx.lineTo(8, -12);
    ctx.quadraticCurveTo(14, -9, 13, -3);
    ctx.lineTo(13, 3);
    ctx.quadraticCurveTo(14, 9, 8, 12);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
    ctx.restore();
  }

  game(state) {
    const {
      width,
      height,
      game,
      buttons,
      cardRects,
      hintIndexes,
      removedAnimations,
      gamePaused,
      toolAdDialog,
      settings
    } = state;
    const ctx = this.ctx;
    drawImageCover(ctx, this.assets.image("gameBg"), 0, 0, width, height);
    this.gameRoundButton(buttons.back, "pause");
    this.gameStats(width, height, game);
    this.gameBoardPanel(width, height);

    for (let index = 0; index < cardRects.length; index += 1) {
      const rect = cardRects[index];
      const card = game.cards[index];
      if (!card || card.state === "removed") {
        continue;
      }
      const hinted = hintIndexes.includes(index);
      this.card(rect, card, hinted);
    }

    for (const animation of removedAnimations || []) {
      this.removedCardAnimation(animation);
    }

    this.gameTools(width, buttons, state.progress.hints, state.previewAgainCount, state.removePairCount);
    this.gameInstruction(width, height);

    if (game.state === "previewing_cards") {
      this.toast("\u8bb0\u4f4f\u5b83\u4eec\u7684\u4f4d\u7f6e", width, height);
    }

    if (gamePaused) {
      if (toolAdDialog) {
        this.toolAdDialog(width, height, buttons, toolAdDialog);
      } else {
        this.pauseMenu(width, height, buttons, settings);
      }
      return;
    }

    if (game.isEnded()) {
      this.resultOverlay(state);
    }
  }

  pauseMenu(width, height, buttons, settings) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.38)";
    ctx.fillRect(0, 0, width, height);

    const panelW = Math.min(width * 0.78, 320);
    const panelH = 280;
    const x = (width - panelW) / 2;
    const y = height * 0.27;
    ctx.fillStyle = "#fff9df";
    roundRect(ctx, x, y, panelW, panelH, 18);
    ctx.fill();
    ctx.strokeStyle = "#c8842c";
    ctx.lineWidth = 4;
    ctx.stroke();

    strokeTextCenter(ctx, "\u64cd\u4f5c", x, y + 36, panelW, 24, "#ffffff", "#8f4f17", 4, "900");
    this.closeIconButton(buttons.pauseClose);
    this.settingsSwitch(buttons.pauseMusic, "\u97f3\u4e50", settings.musicEnabled);
    this.settingsSwitch(buttons.pauseSfx, "\u97f3\u6548", settings.sfxEnabled);
    this.dialogButton(buttons.pauseRetry, "\u91cd\u73a9", "#f5b43c", "#b96a09");
    this.dialogButton(buttons.pauseHome, "\u9996\u9875", "#78d64b", "#4ba733");
    ctx.restore();
  }

  toolAdDialog(width, height, buttons, dialog) {
    const ctx = this.ctx;
    ctx.save();
    ctx.fillStyle = "rgba(0, 0, 0, 0.38)";
    ctx.fillRect(0, 0, width, height);

    const panelW = Math.min(width * 0.78, 320);
    const panelH = 188;
    const x = (width - panelW) / 2;
    const y = height * 0.31;
    ctx.fillStyle = "#fff9df";
    roundRect(ctx, x, y, panelW, panelH, 18);
    ctx.fill();
    ctx.strokeStyle = "#c8842c";
    ctx.lineWidth = 4;
    ctx.stroke();

    strokeTextCenter(ctx, "\u6b21\u6570\u5df2\u7528\u5b8c", x, y + 36, panelW, 23, "#ffffff", "#8f4f17", 4, "900");
    fillTextCenter(ctx, `${dialog.label}\u6b21\u6570\u5df2\u7528\u5b8c`, x, y + 78, panelW, 16, "#7b3f17", "900");
    fillTextCenter(ctx, "\u89c2\u770b\u5e7f\u544a\u53ef\u589e\u52a01\u6b21", x, y + 104, panelW, 15, "#7b3f17", "800");
    this.dialogButton(buttons.toolAdCancel, "\u5173\u95ed", "#d5d0c4", "#8a8275");
    this.dialogButton(buttons.toolAdConfirm, "\u770b\u5e7f\u544a", "#f5b43c", "#b96a09");
    ctx.restore();
  }

  closeIconButton(rect) {
    const ctx = this.ctx;
    const cx = rect.x + rect.w / 2;
    const cy = rect.y + rect.h / 2;
    ctx.save();
    ctx.fillStyle = "#f5b43c";
    ctx.beginPath();
    ctx.arc(cx, cy, rect.w / 2, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#b96a09";
    ctx.lineWidth = 3;
    ctx.stroke();
    ctx.strokeStyle = "#ffffff";
    ctx.lineWidth = 3.5;
    ctx.lineCap = "round";
    ctx.beginPath();
    ctx.moveTo(cx - rect.w * 0.22, cy - rect.h * 0.22);
    ctx.lineTo(cx + rect.w * 0.22, cy + rect.h * 0.22);
    ctx.moveTo(cx + rect.w * 0.22, cy - rect.h * 0.22);
    ctx.lineTo(cx - rect.w * 0.22, cy + rect.h * 0.22);
    ctx.stroke();
    ctx.restore();
  }

  gameRoundButton(rect, type) {
    const ctx = this.ctx;
    const cx = rect.x + rect.w / 2;
    const cy = rect.y + rect.h / 2;
    const r = rect.w / 2;
    ctx.save();
    ctx.fillStyle = "rgba(93, 54, 18, 0.25)";
    ctx.beginPath();
    ctx.arc(cx + 2, cy + 3, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = "#a86425";
    ctx.beginPath();
    ctx.arc(cx, cy, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#fff0bf";
    ctx.lineWidth = 3;
    ctx.stroke();
    ctx.fillStyle = "#ffffff";
    if (type === "pause") {
      roundRect(ctx, cx - r * 0.28, cy - r * 0.36, r * 0.18, r * 0.72, 3);
      ctx.fill();
      roundRect(ctx, cx + r * 0.1, cy - r * 0.36, r * 0.18, r * 0.72, 3);
      ctx.fill();
    } else {
      ctx.translate(cx, cy);
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = Math.max(3, r * 0.16);
      for (let index = 0; index < 8; index += 1) {
        ctx.rotate(Math.PI / 4);
        ctx.beginPath();
        ctx.moveTo(0, -r * 0.25);
        ctx.lineTo(0, -r * 0.52);
        ctx.stroke();
      }
      ctx.beginPath();
      ctx.arc(0, 0, r * 0.24, 0, Math.PI * 2);
      ctx.stroke();
    }
    ctx.restore();
  }

  gameTitle(width, height) {
    const ctx = this.ctx;
    const y = height * 0.07;
    strokeTextCenter(ctx, "\u7ffb\u724c\u6d88\u9664", width * 0.17, y + 34, width * 0.66, Math.max(34, width * 0.12), "#fff8df", "#b96a09", 7, "900");
    ctx.save();
    ctx.fillStyle = "#f5c15c";
    roundRect(ctx, width * 0.32, y + 76, width * 0.36, 28, 5);
    ctx.fill();
    ctx.strokeStyle = "#d89436";
    ctx.lineWidth = 2;
    ctx.stroke();
    fillTextCenter(ctx, "\u7ffb\u5f00\u76f8\u540c\u56fe\u6848\u8fdb\u884c\u6d88\u9664\uff01", width * 0.29, y + 90, width * 0.42, 15, "#7b3f17", "900");
    ctx.restore();
  }

  gameStats(width, height, game) {
    const y = height * 0.13;
    const w = width * 0.27;
    const h = 58;
    const items = [{ label: "\u7b49\u7ea7", value: `1-${game.level.levelId}`, icon: "flower" }];
    if (game.level.showTimer !== false) {
      items.push({ label: "\u65f6\u95f4", value: this.formatTime(game.remainingSeconds()), icon: "clock" });
    }
    if (game.level.showSteps !== false) {
      items.push({ label: "\u6b65\u6570", value: String(game.steps), icon: "star" });
    }
    if (game.level.showMismatch !== false) {
      items.push({ label: "\u5931\u8bef", value: `${game.mismatchCount}/${game.level.maxMismatchCount}`, icon: "flower" });
    }

    const gap = width * 0.02;
    const itemW = Math.min(w, (width * 0.84 - gap * (items.length - 1)) / items.length);
    const startX = (width - itemW * items.length - gap * (items.length - 1)) / 2;
    items.forEach((item, index) => {
      this.gameStatCard(startX + index * (itemW + gap), y, itemW, h, item.label, item.value, item.icon);
    });
  }

  gameStatCard(x, y, w, h, label, value, icon) {
    const ctx = this.ctx;
    const isClock = icon === "clock";
    const iconSize = isClock ? 9 : 11;
    const iconX = x + (isClock ? w * 0.12 : w * 0.18);
    const iconY = y + 39;
    const valueLeft = x + (isClock ? w * 0.38 : w * 0.34);
    const valueWidth = Math.max(18, w - (valueLeft - x) - w * 0.08);
    ctx.save();
    ctx.fillStyle = "rgba(104, 64, 20, 0.15)";
    roundRect(ctx, x + 2, y + 3, w, h, 15);
    ctx.fill();
    ctx.fillStyle = "#fff8e7";
    roundRect(ctx, x, y, w, h, 15);
    ctx.fill();
    ctx.strokeStyle = "#ecd19a";
    ctx.lineWidth = 2;
    ctx.stroke();
    fillTextCenter(ctx, label, x, y + 17, w, 15, "#7b3f17", "900");
    if (icon === "star") {
      this.starPath(iconX, iconY, iconSize, "#ffd84d", "#d98d22");
    } else if (icon === "clock") {
      ctx.fillStyle = "#ffffff";
      ctx.strokeStyle = "#2f9add";
      ctx.lineWidth = 3;
      ctx.beginPath();
      ctx.arc(iconX, iconY, iconSize, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
      ctx.fillStyle = "#2f9add";
      ctx.beginPath();
      ctx.arc(iconX, iconY, 2.2, 0, Math.PI * 2);
      ctx.fill();
      ctx.strokeStyle = "#2f9add";
      ctx.lineCap = "round";
      ctx.lineWidth = 2.5;
      ctx.beginPath();
      ctx.moveTo(iconX, iconY);
      ctx.lineTo(iconX, iconY - 5.2);
      ctx.moveTo(iconX, iconY);
      ctx.lineTo(iconX + 4.3, iconY + 2.6);
      ctx.stroke();
    } else {
      this.starPath(iconX, iconY, iconSize, "#9bdc59", "#65aa36");
    }
    fillTextCenter(ctx, value, valueLeft, y + 39, valueWidth, isClock ? 19 : 24, "#7b3f17", "900");
    ctx.restore();
  }

  formatTime(seconds) {
    const minute = Math.floor(seconds / 60);
    const second = seconds % 60;
    return `${String(minute).padStart(2, "0")}:${String(second).padStart(2, "0")}`;
  }

  gameBoardPanel(width, height) {
    const ctx = this.ctx;
    const x = width * 0.045;
    const y = height * 0.225;
    const w = width * 0.91;
    const h = height * 0.44;
    ctx.save();
    ctx.fillStyle = "rgba(100, 58, 17, 0.16)";
    roundRect(ctx, x + 2, y + 5, w, h, 24);
    ctx.fill();
    ctx.fillStyle = "#fff4dc";
    roundRect(ctx, x, y, w, h, 24);
    ctx.fill();
    ctx.strokeStyle = "#e7c27a";
    ctx.lineWidth = 3;
    ctx.stroke();
    ctx.restore();
  }

  gameTools(width, buttons, hints, previewAgainCount, removePairCount) {
    const ctx = this.ctx;
    const trayX = width * 0.18;
    const trayY = buttons.hint.y - 8;
    const trayW = width * 0.64;
    const trayH = 88;
    ctx.save();
    ctx.fillStyle = "#d99545";
    roundRect(ctx, trayX, trayY, trayW, trayH, 20);
    ctx.fill();
    ctx.strokeStyle = "#b8732c";
    ctx.lineWidth = 3;
    ctx.stroke();
    this.gameToolButton(buttons.hint, "\u63d0\u793a", hints, "toolBulb");
    this.gameToolButton(buttons.shuffle, "\u518d\u770b\u4e00\u6b21", previewAgainCount, "toolMagnifier");
    this.gameToolButton(buttons.pair, "\u6d88\u9664\u4e00\u5bf9", removePairCount, "toolEraser");
    ctx.restore();
  }

  gameToolButton(rect, label, count, type) {
    const ctx = this.ctx;
    const cx = rect.x + rect.w / 2;
    const cy = rect.y + rect.h * 0.36;
    const r = Math.min(rect.w, rect.h) * 0.36;
    ctx.save();
    ctx.fillStyle = "#8be0c5";
    ctx.beginPath();
    ctx.arc(cx, cy, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#fff6db";
    ctx.lineWidth = 4;
    ctx.stroke();
    const iconSize = r * 1.18;
    drawImageContain(ctx, this.assets.image(type), cx - iconSize / 2, cy - iconSize / 2, iconSize, iconSize);
    ctx.fillStyle = "#e78018";
    const badgeX = cx + r * 0.54;
    const badgeY = cy - r * 0.54;
    const badgeR = r * 0.34;
    ctx.beginPath();
    ctx.arc(badgeX, badgeY, badgeR, 0, Math.PI * 2);
    ctx.fill();
    strokeTextCenter(ctx, String(count), badgeX - badgeR, badgeY, badgeR * 2, 15, "#ffffff", "#b95d08", 2, "900");
    ctx.fillStyle = "#fff8e6";
    roundRect(ctx, rect.x + rect.w * 0.08, rect.y + rect.h * 0.72, rect.w * 0.84, 22, 10);
    ctx.fill();
    fillTextCenter(ctx, label, rect.x, rect.y + rect.h * 0.91, rect.w, 14, "#7b3f17", "900");
    ctx.restore();
  }

  gameInstruction(width, height) {
    const ctx = this.ctx;
    const x = width * 0.17;
    const y = height * 0.875;
    const w = width * 0.66;
    const h = 48;
    ctx.save();
    ctx.fillStyle = "#fff4dc";
    roundRect(ctx, x, y, w, h, 16);
    ctx.fill();
    ctx.strokeStyle = "#e4bd75";
    ctx.lineWidth = 3;
    ctx.stroke();
    this.starPath(x + 30, y + h / 2, 12, "#ffd84d", "#d98d22");
    fillTextCenter(ctx, "\u7ffb\u5f00\u4e24\u5f20\u76f8\u540c\u7684\u56fe\u6848\u5373\u53ef\u6d88\u9664", x + 42, y + h / 2, w - 52, 13, "#7b3f17", "900");
    ctx.restore();
  }

  header(title, width, backButton) {
    const ctx = this.ctx;
    ctx.fillStyle = "rgba(255,255,255,0.86)";
    roundRect(ctx, 16, 18, width - 32, 46, 10);
    ctx.fill();
    drawImageContain(ctx, this.assets.image("back"), backButton.x, backButton.y, backButton.w, backButton.h);
    fillTextCenter(ctx, title, 0, 42, width, 22, "#23455c");
  }

  drawButton(rect, text, color) {
    const ctx = this.ctx;
    ctx.fillStyle = color;
    roundRect(ctx, rect.x, rect.y, rect.w, rect.h, 10);
    ctx.fill();
    ctx.strokeStyle = "rgba(89,71,36,0.22)";
    ctx.lineWidth = 2;
    ctx.stroke();
    fillTextCenter(ctx, text, rect.x, rect.y + rect.h / 2, rect.w, 16, "#4b3810");
  }

  statusText(text, x, y, color) {
    const ctx = this.ctx;
    ctx.fillStyle = color;
    ctx.font = "700 16px sans-serif";
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillText(text, x, y);
  }

  card(rect, card, hinted) {
    const ctx = this.ctx;
    ctx.save();
    if (hinted) {
      ctx.shadowColor = "#ffdf66";
      ctx.shadowBlur = 18;
    }

    ctx.fillStyle = "#ffffff";
    roundRect(ctx, rect.x, rect.y, rect.w, rect.h, 10);
    ctx.fill();
    ctx.strokeStyle = hinted ? "#ffbf2f" : "#5aa7be";
    ctx.lineWidth = hinted ? 4 : 2;
    ctx.stroke();

    if (card.state === "face_down") {
      drawImageContain(ctx, this.assets.image("cardBack"), rect.x + 5, rect.y + 5, rect.w - 10, rect.h - 10);
    } else {
      drawImageContain(ctx, this.assets.image(`animal:${card.iconKey}`), rect.x + 10, rect.y + 10, rect.w - 20, rect.h - 20);
    }

    ctx.restore();
  }

  removedCardAnimation(animation) {
    const ctx = this.ctx;
    const rect = animation.rect;
    const duration = Math.max(1, animation.endAt - animation.startAt);
    const progress = Math.max(0, Math.min(1, (Date.now() - animation.startAt) / duration));
    const scale = 1 - progress * 0.22;
    const alpha = 1 - progress;
    const w = rect.w * scale;
    const h = rect.h * scale;
    const x = rect.x + (rect.w - w) / 2;
    const y = rect.y + (rect.h - h) / 2;

    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.shadowColor = "rgba(255, 221, 92, 0.75)";
    ctx.shadowBlur = 18 * (1 - progress);
    ctx.fillStyle = "#ffffff";
    roundRect(ctx, x, y, w, h, 10);
    ctx.fill();
    ctx.strokeStyle = "#76d95a";
    ctx.lineWidth = 3;
    ctx.stroke();
    drawImageContain(ctx, this.assets.image(`animal:${animation.iconKey}`), x + w * 0.1, y + h * 0.1, w * 0.8, h * 0.8);
    ctx.restore();
  }

  toast(text, width, height) {
    const ctx = this.ctx;
    const boxW = Math.min(width * 0.58, 250);
    const boxH = 40;
    const x = (width - boxW) / 2;
    const y = height * 0.205;
    ctx.save();
    ctx.fillStyle = "rgba(41, 84, 93, 0.86)";
    roundRect(ctx, x, y, boxW, boxH, boxH / 2);
    ctx.fill();
    ctx.strokeStyle = "rgba(255, 255, 255, 0.5)";
    ctx.lineWidth = 2;
    ctx.stroke();
    strokeTextCenter(ctx, text, x, y + boxH / 2, boxW, 17, "#ffffff", "rgba(24, 61, 70, 0.9)", 3, "900");
    ctx.restore();
  }

  resultOverlay(state) {
    const { width, height, game, buttons } = state;
    const ctx = this.ctx;
    const result = game.lastResult;
    ctx.fillStyle = "rgba(0,0,0,0.42)";
    ctx.fillRect(0, 0, width, height);

    const panelW = width * 0.78;
    const panelH = 260;
    const x = (width - panelW) / 2;
    const y = height * 0.28;
    ctx.fillStyle = "#ffffff";
    roundRect(ctx, x, y, panelW, panelH, 16);
    ctx.fill();

    fillTextCenter(ctx, result.success ? "\u95ef\u5173\u6210\u529f" : "\u6311\u6218\u5931\u8d25", x, y + 46, panelW, 28, "#23455c");
    if (result.success) {
      const size = 34;
      for (let index = 0; index < 3; index += 1) {
        const image = index < result.stars ? this.assets.image("winStar") : this.assets.image("lossStar");
        drawImageContain(ctx, image, x + panelW / 2 - 57 + index * 40, y + 74, size, size);
      }
      fillTextCenter(ctx, `\u5956\u52b1 ${result.coinsEarned} \u91d1\u5e01`, x, y + 132, panelW, 18, "#4f6f7d", "600");
      if (result.rewards && result.rewards.stamina > 0) {
        fillTextCenter(ctx, `\u9996\u6b213\u661f +${result.rewards.stamina}\u4f53\u529b`, x, y + 154, panelW, 17, "#2d9a58", "700");
      }
    } else {
      fillTextCenter(ctx, this.reasonText(result.reason), x, y + 102, panelW, 18, "#4f6f7d", "600");
    }
    fillTextCenter(ctx, `\u7528\u65f6 ${Math.ceil(result.elapsedMs / 1000)}s  \u6b65\u6570 ${result.steps}`, x, y + 166, panelW, 17, "#4f6f7d", "600");
    this.drawButton(buttons.retry, "\u91cd\u73a9", "#f7b84b");
    this.drawButton(buttons.exit, "\u9000\u51fa", "#d5d0c4");
    this.drawButton(buttons.next, result.success ? "\u4e0b\u4e00\u5173" : "\u9009\u5173", "#76d6bd");
  }

  reasonText(reason) {
    if (reason === "time_out") {
      return "\u65f6\u95f4\u7528\u5b8c\u4e86";
    }
    if (reason === "mismatch_limit") {
      return "\u9519\u8bef\u6b21\u6570\u7528\u5b8c\u4e86";
    }
    return "\u518d\u8bd5\u4e00\u6b21";
  }
}

module.exports = {
  Renderer,
  clamp
};

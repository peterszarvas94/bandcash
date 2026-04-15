const paddleSelector = "[data-paddle-price-id]";
let paddleInitialized = false;
let paddleInitInFlight = false;

function getPaddleConfig() {
  const body = document.body;
  if (!body) {
    return { env: "", token: "" };
  }

  return {
    env: (body.dataset.paddleEnv || "").trim(),
    token: (body.dataset.paddleClientToken || "").trim(),
  };
}

async function waitForPaddle(maxWaitMs) {
  const start = Date.now();
  while (Date.now() - start < maxWaitMs) {
    if (window.Paddle && window.Paddle.Initialize && window.Paddle.Checkout) {
      return window.Paddle;
    }
    await new Promise((resolve) => {
      window.setTimeout(resolve, 50);
    });
  }
  return null;
}

async function ensurePaddleInitialized() {
  if (paddleInitialized) {
    return true;
  }
  if (paddleInitInFlight) {
    return false;
  }

  const config = getPaddleConfig();
  if (!config.env || !config.token) {
    return false;
  }

  paddleInitInFlight = true;
  try {
    const paddle = await waitForPaddle(6000);
    if (!paddle) {
      return false;
    }
    if (config.env === "sandbox" && paddle.Environment && paddle.Environment.set) {
      paddle.Environment.set("sandbox");
    }
    paddle.Initialize({ token: config.token });
    paddleInitialized = true;
    return true;
  } finally {
    paddleInitInFlight = false;
  }
}

async function onCheckoutClick(event) {
  const button = event.target.closest(paddleSelector);
  if (!button) {
    return;
  }
  const priceId = (button.getAttribute("data-paddle-price-id") || "").trim();
  const userID = (button.getAttribute("data-paddle-user-id") || "").trim();
  const userEmail = (button.getAttribute("data-paddle-user-email") || "").trim();
  if (!priceId) {
    return;
  }

  event.preventDefault();
  const ready = await ensurePaddleInitialized();
  if (!ready || !window.Paddle || !window.Paddle.Checkout) {
    return;
  }

  const checkoutInput = {
    items: [
      {
        priceId,
      },
    ],
  };

  if (userID) {
    checkoutInput.customData = { user_id: userID };
  }
  if (userEmail) {
    checkoutInput.customer = { email: userEmail };
  }

  window.Paddle.Checkout.open(checkoutInput);
}

document.addEventListener("click", onCheckoutClick);

# Trading Flow

- All these operations are on the trading pair declared in the environment configuration.
- If an error occurs, notify the user.

## 1. Get Price

- Get the price.

## 2. Get Open Position

- Get the last buy of the trade.

## 3. Calculate Potential Profit

- Calculate the potential profit.
- If the potential profit is less than **-5%**, trigger an emergency **SELL** by setting the signal to **SELL**.

## 4. Analyze Market Data

- Analyze the market data based on the trading strategy.
- If the signal is not `nil`, set the signal to the strategy signal.
- The signal is determined based on the market strategy, e.g., mean reversion following the RSI indicator.
- Log the signal.

## 5. Execute Signal

### If the signal is **BUY**:

- Get the current balance.
- Calculate the position size based on USDT balance and risk management.
- If the position size is greater than the max quantity, set the position size to the max quantity.
- Ensure the minimum order size.
- Generate a position ID for tracking.
- Place the **BUY** order.
- Save the trade to the database.
- Notify the user.

### If the signal is **SELL**:

- Execute the **SELL** order.
- Check the last **BUY** trade. If not `nil`, then calculate the potential profit.
- If the potential profit is less than **-8%**, set the signal to **SELL**.
- Add protection to sell the position if the potential profit is less than **-8%**.
- **Condition to sell:**
  - Get a **SELL** signal.
  - Have crypto balance to sell.
  - Potential profit is greater than **0%**.
  - Potential profit is greater than **2%**.
- **Tiered exit system:**
  - Sell **50%** at **5%** profit.
  - Sell **30%** at **3%** profit.
- Place the **SELL** order.
- Update the original **BUY** trade status to **CLOSED** and save the trade to the database with proper position linking.
- Notify the user.

## 6. Adjust Trade Frequency

- Can adjust trade frequency based on strategy and market conditions.

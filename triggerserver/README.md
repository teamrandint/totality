# TRIGGER SERVER SPEC

## ENDPOINTS

### SET_BUY_TRIGGER

params: transnum, user, stock, price

returns: success or not

### CANCEL_SET_BUY

params: transnum, user, stock

returns: success or not

### SET_SELL_TRIGGER

params: transnum, user, stock, price

returns: success or not

### CANCEL_SET_SELL

params: transnum, user, stock

returns: success or not

## TRIGGER OBJECT SPEC

- username
- stockname
- price
- action
  - buy or sell

## BEHAVIOUR

For each running trigger, the server polls the quoteserver at creation time. Since this will cache a quote for 60s, the trigger will sleep for 60s.

If, the trigger is successful at any polling:

- Log the success
- Stop the poll loop
- Hit the transaction server to perform the reserve account transactions
- Close the trigger
- Log the trigger being close

## IMPLEMENTATION REQUIRED

- Implement a trigger server to do this BEHAVIOUR, responding to ENDPOINTS
- Add database functionality for the user+":SellTriggers" and user+":BuyTriggers" fields
- Rewrite the trigger library in the transaction server to hit this servers endpoints, as well as the new database endpoints

DROP TABLE IF EXISTS transactionsDefault;

CREATE TABLE transactionsDefault (
  id UUID PRIMARY KEY,
  amount FLOAT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DROP TABLE IF EXISTS transactionsFallback;

CREATE TABLE transactionsFallback (
  id UUID PRIMARY KEY,
  amount FLOAT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
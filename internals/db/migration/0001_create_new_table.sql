
CREATE EXTENSION IF NO EXISTS "uuid-ossp";


--Tables--- 
-- CREATE TABLE crews(
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     name VARCHAR(100) NOT NULL,
--     owner_id UUID NOT NULL, 
--     created_at TIMESTAMPZ NOT NULL DEFAULT NOW(),
--     FOREIGN KEY (owner_id) REFERENCE users(id) ON DELETE CASCADE
-- )



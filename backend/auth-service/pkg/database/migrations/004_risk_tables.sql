-- +migrate Up
-- Risk Categories table
CREATE TABLE IF NOT EXISTS risk_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Likelihoods table
CREATE TABLE IF NOT EXISTS likelihoods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    score INT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Impacts table
CREATE TABLE IF NOT EXISTS impacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    score INT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Risk Levels table
CREATE TABLE IF NOT EXISTS risk_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    min_score INT NOT NULL,
    max_score INT NOT NULL,
    color VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Risk Matrix Cells table (likelihood x impact matrix)
CREATE TABLE IF NOT EXISTS risk_matrix_cells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    likelihood_id UUID NOT NULL REFERENCES likelihoods(id) ON DELETE CASCADE,
    impact_id UUID NOT NULL REFERENCES impacts(id) ON DELETE CASCADE,
    score INT NOT NULL,
    risk_level_id UUID NOT NULL REFERENCES risk_levels(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(likelihood_id, impact_id)
);

-- Risk Register table
CREATE TABLE IF NOT EXISTS risk_registers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    department_id UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    risk_category_id UUID NOT NULL REFERENCES risk_categories(id) ON DELETE CASCADE,
    risk_owner_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,

    -- Inherent Risk (before controls)
    inherent_likelihood_id UUID NOT NULL REFERENCES likelihoods(id) ON DELETE CASCADE,
    inherent_impact_id UUID NOT NULL REFERENCES impacts(id) ON DELETE CASCADE,
    inherent_score INT NOT NULL,
    inherent_risk_level_id UUID NOT NULL REFERENCES risk_levels(id) ON DELETE CASCADE,

    -- Residual Risk (after controls)
    residual_likelihood_id UUID NOT NULL REFERENCES likelihoods(id) ON DELETE CASCADE,
    residual_impact_id UUID NOT NULL REFERENCES impacts(id) ON DELETE CASCADE,
    residual_score INT NOT NULL,
    residual_risk_level_id UUID NOT NULL REFERENCES risk_levels(id) ON DELETE CASCADE,

    assessment_date TIMESTAMP WITH TIME ZONE,
    next_review_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- +migrate Down
DROP TABLE IF EXISTS risk_registers CASCADE;
DROP TABLE IF EXISTS risk_matrix_cells CASCADE;
DROP TABLE IF EXISTS risk_levels CASCADE;
DROP TABLE IF EXISTS impacts CASCADE;
DROP TABLE IF EXISTS likelihoods CASCADE;
DROP TABLE IF EXISTS risk_categories CASCADE;

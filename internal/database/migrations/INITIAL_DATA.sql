CREATE TABLE IF NOT EXISTS users (
  id         serial       NOT NULL,
  username   VARCHAR(150) NOT NULL,
  password   VARCHAR(256) NOT NULL,
  created_at timestamp    DEFAULT now(),

  CONSTRAINT pk_users PRIMARY KEY(id)
);

INSERT INTO users (id, username, password)
  VALUES (1, 'superadmin', '$2a$10$usFqeLL8Z3xLEznqYCYdL.c9V3uO3odB2Ub8AWwOITUjvxE6iUBuW')
  ON CONFLICT(id) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
  id          serial       NOT NULL,
  name        VARCHAR(150) NOT NULL,
  description VARCHAR(256) NOT NULL,
  deletable   bool         DEFAULT TRUE,
  created_at  timestamp    DEFAULT NOW(),
  updated_at  timestamp    DEFAULT NOW(),

  CONSTRAINT pk_permissions PRIMARY KEY(id)
);

INSERT INTO permissions (id, name, description, deletable)
  VALUES
    (1, 'create_users', 'Poder crear usuarios dentro de la aplicación', FALSE),
    (2, 'delete_users', 'Poder eliminar usuarios dentro de la aplicación', FALSE),
    (3, 'create_permissions', 'Poder crear permisos para la aplicación', FALSE),
    (4, 'update_permission', 'Poder actualizar un permiso de la aplicación', FALSE),
    (5, 'delete_permissions', 'Poder eliminar permisos de la aplicación', FALSE)
  ON CONFLICT(id) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_permissions (
  id            serial NOT NULL,
  user_id       serial NOT NULL,
  permission_id serial NOT NULL,

  CONSTRAINT pk_user_permissions PRIMARY KEY(id),
  CONSTRAINT fk_user_permissions_uid FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_permissions_pid FOREIGN KEY(permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

INSERT INTO user_permissions (id, user_id, permission_id)
  VALUES (1, 1, 1), (2, 1, 2), (3, 1, 3), (4, 1, 4), (5, 1, 5)
  ON CONFLICT(id) DO NOTHING;
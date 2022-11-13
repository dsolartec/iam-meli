CREATE TABLE IF NOT EXISTS users (
  id         serial       NOT NULL,
  username   VARCHAR(10)  NOT NULL,
  password   VARCHAR(256) NOT NULL,
  created_at timestamp    DEFAULT now(),

  CONSTRAINT pk_users PRIMARY KEY(id)
);

INSERT INTO users (id, username, password)
  VALUES (1, 'superadmin', '$2a$10$usFqeLL8Z3xLEznqYCYdL.c9V3uO3odB2Ub8AWwOITUjvxE6iUBuW')
  ON CONFLICT(id) DO NOTHING;

SELECT SETVAL(
  (SELECT pg_get_serial_sequence('users', 'id')),
  (SELECT MAX(id) FROM users)
);

CREATE TABLE IF NOT EXISTS permissions (
  id          SERIAL       NOT NULL,
  name        VARCHAR(25)  NOT NULL,
  description VARCHAR(150) NOT NULL,
  deletable   BOOLEAN      DEFAULT TRUE,
  editable    BOOLEAN      DEFAULT TRUE,
  created_at  timestamp    DEFAULT NOW(),
  updated_at  timestamp    DEFAULT NOW(),

  CONSTRAINT pk_permissions PRIMARY KEY(id)
);

INSERT INTO permissions (id, name, description, deletable, editable)
  VALUES
    (1, 'delete_user', 'Poder eliminar usuarios dentro de la aplicación', FALSE, FALSE),
    (2, 'create_permission', 'Poder crear permisos para la aplicación', FALSE, FALSE),
    (3, 'update_permission', 'Poder actualizar un permiso de la aplicación', FALSE, FALSE),
    (4, 'grant_permission', 'Poder añadir un permiso a un usuario', FALSE, FALSE),
    (5, 'revoke_permission', 'Poder eliminar un permiso a un usuario', FALSE, FALSE),
    (6, 'delete_permission', 'Poder eliminar permisos de la aplicación', FALSE, FALSE)
  ON CONFLICT(id) DO NOTHING;

SELECT SETVAL(
  (SELECT pg_get_serial_sequence('permissions', 'id')),
  (SELECT MAX(id) FROM permissions)
);

CREATE TABLE IF NOT EXISTS user_permissions (
  id            serial NOT NULL,
  user_id       serial NOT NULL,
  permission_id serial NOT NULL,

  CONSTRAINT pk_user_permissions PRIMARY KEY(id),
  CONSTRAINT fk_user_permissions_uid FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_permissions_pid FOREIGN KEY(permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

INSERT INTO user_permissions (id, user_id, permission_id)
  VALUES (1, 1, 1), (2, 1, 2), (3, 1, 3), (4, 1, 4), (5, 1, 5), (6, 1, 6)
  ON CONFLICT(id) DO NOTHING;

SELECT SETVAL(
  (SELECT pg_get_serial_sequence('user_permissions', 'id')),
  (SELECT MAX(id) FROM user_permissions)
);

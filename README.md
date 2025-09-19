# Google Auth Wizard 🧙‍♂️

Un outil en ligne de commande interactif pour simplifier l'authentification OAuth2 avec les APIs Google. Sélectionnez facilement les scopes Google nécessaires via une interface terminal moderne et obtenez vos tokens d'accès.

## ✨ Fonctionnalités

- 🔍 **Interface interactive** : Sélection des scopes via une interface terminal intuitive (powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea))
- 🌐 **Récupération automatique des scopes** : Fetch automatique des scopes disponibles depuis Google OAuth Playground
- 🔧 **Configuration flexible** : Configuration via fichier YAML avec valeurs par défaut
- 🚀 **Serveur OAuth temporaire** : Serveur local automatique pour le callback OAuth
- 📋 **Gestion d'erreurs robuste** : Gestion gracieuse des erreurs avec messages informatifs
- 🎨 **Interface moderne** : Styling avec couleurs et navigation au clavier

## 📦 Installation

### Prérequis

- Go 1.19+ 
- Un projet Google Cloud avec OAuth2 configuré
- Un fichier client secret JSON de Google Cloud

### Compilation

```bash
git clone https://github.com/votre-username/google-auth-wizard
cd google-auth-wizard
go mod download
go build -o google-auth-wizard
```

## 🚀 Utilisation

### Commande de base

```bash
./google-auth-wizard -file client_secret_[ID].apps.googleusercontent.com.json
```

ou

```bash
./google-auth-wizard -f client_secret_[ID].apps.googleusercontent.com.json
```

### Avec Go Run (développement)

```bash
go run main.go -file client_secret_[ID].apps.googleusercontent.com.json
```

### Workflow typique

1. **Lancement** : Exécutez la commande avec votre fichier client secret
2. **Sélection des APIs** : Naviguez et sélectionnez les services Google APIs 
3. **Sélection des scopes** : Choisissez les scopes spécifiques pour chaque service
4. **Authentification** : Le navigateur s'ouvre automatiquement pour l'OAuth
5. **Récupération du token** : Le token d'accès est affiché dans le terminal

### Navigation

- `↑`/`↓` : Navigation dans les listes
- `Espace` : Sélection/désélection des items
- `Entrée` : Confirmer la sélection
- `Esc` : Retour au niveau précédent
- `q` : Quitter l'application

## ⚙️ Configuration

### Fichier config.yaml

Le fichier de configuration est créé automatiquement avec des valeurs par défaut :

```yaml
# Google Auth Wizard Configuration
server:
  # Port par défaut pour le serveur de callback OAuth
  defaultPort: 8080
  
  # Nombre maximum de ports à essayer si le port par défaut est occupé
  maxPortTries: 10
  
  # Timeout pour le serveur de callback OAuth
  serverTimeout: 5m0s

oauth:
  # Chemin de callback OAuth
  callbackPath: /callback
  
  # URL Google OAuth playground pour récupérer les scopes
  oauthPlaygroundURL: https://developers.google.com/oauthplayground
  
  # Endpoint pour récupérer les scopes
  scopeEndpoint: getScopes
  
  # Timeout pour les requêtes de récupération des scopes
  scopeTimeout: 1m0s

terminal:
  # Hauteur de l'interface terminal (nombre d'items affichés)
  height: 20
```

### Personnalisation

Vous pouvez modifier le fichier `config.yaml` pour ajuster :
- Les ports utilisés pour le serveur OAuth
- Les timeouts de connexion
- La hauteur de l'interface terminal
- Les URLs des endpoints Google

## 🏗️ Architecture

```
google-auth-wizard/
├── main.go              # Point d'entrée principal
├── auth/
│   └── oauth.go         # Logique d'authentification OAuth2
├── config/
│   └── config.go        # Gestion de la configuration
├── googlescopes/
│   └── client.go        # Client pour récupérer les scopes Google
├── terminal/
│   ├── terminal.go      # Interface utilisateur terminal
│   └── struct.go        # Structures de données UI
├── utils/
│   └── utils.go         # Utilitaires (parsing, navigation, ports)
├── config.yaml          # Configuration
└── go.mod               # Dépendances Go
```

## 🔧 Développement

### Tests

```bash
go test ./...
```

### Build pour production

```bash
go build -ldflags="-s -w" -o google-auth-wizard
```

### Nettoyage

```bash
go mod tidy
```

## 🤝 Contribution

Les contributions sont les bienvenues ! Veuillez :

1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/nouvelle-fonctionnalite`)
3. Commit vos changements (`git commit -am 'Ajout nouvelle fonctionnalite'`)
4. Push sur la branche (`git push origin feature/nouvelle-fonctionnalite`)
5. Ouvrir une Pull Request

## 📝 Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

## 🆘 Support

- 📚 [Documentation Google OAuth2](https://developers.google.com/identity/protocols/oauth2)
- 🔧 [Configuration Google Cloud Console](https://console.cloud.google.com/)
- 💬 Ouvrez une issue pour signaler un bug ou demander une fonctionnalité

## 🏷️ Versions

### v1.0.0
- Interface terminal interactive
- Récupération automatique des scopes
- Configuration YAML
- Gestion d'erreurs améliorée
- Architecture modulaire

---

Développé avec ❤️ en Go
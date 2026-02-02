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

- Go 1.21+ 
- Un projet Google Cloud avec OAuth2 configuré
- Créer vos identifiants OAuth2 : https://console.cloud.google.com/apis/credentials?hl=fr&project={{votre nom de project}}
- Activer les API et services nécessaires : https://console.cloud.google.com/apis/dashboard?hl=fr&project={{votre nom de project}}
- Un fichier client secret JSON de Google Cloud

### Compilation

```bash
git clone https://github.com/votre-username/google-auth-wizard
cd google-auth-wizard
go mod download
go build -o google-auth-wizard
```

## 🔐 Configuration Google Cloud

### 1. Créer un projet Google Cloud

1. Accédez à [Google Cloud Console](https://console.cloud.google.com/)
2. Cliquez sur le sélecteur de projet en haut de la page
3. Cliquez sur "Nouveau projet"
4. Donnez un nom à votre projet et cliquez sur "Créer"

### 2. Configurer l'écran de consentement OAuth

Avant de créer des identifiants, vous devez configurer l'écran de consentement :

1. Allez sur [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent)
2. Sélectionnez le type d'utilisateur :
   - **Externe** : Pour tester avec n'importe quel compte Google (recommandé pour le développement)
   - **Interne** : Uniquement si vous avez un compte Google Workspace
3. Remplissez les informations requises :
   - **Nom de l'application** : Le nom qui apparaîtra aux utilisateurs
   - **E-mail de l'utilisateur assistance** : Votre adresse e-mail
   - **Domaines autorisés** : Laissez vide pour les tests
4. Cliquez sur "Enregistrer et continuer"
5. **Scopes** : Ignorez cette section (les scopes seront demandés dynamiquement)
6. **Utilisateurs test** : Ajoutez les adresses e-mail qui pourront tester votre application
   - ⚠️ Important : Si votre app n'est pas publiée, seuls les utilisateurs test pourront se connecter
7. Cliquez sur "Enregistrer et continuer"

### 3. Activer les APIs nécessaires

1. Accédez au [tableau de bord des APIs](https://console.cloud.google.com/apis/dashboard?hl=fr&project={{votre nom de project}})
2. Cliquez sur "+ ACTIVER DES API ET DES SERVICES"
3. Recherchez et activez les APIs dont vous avez besoin (exemple : Google Drive API, Gmail API, etc.)
4. Répétez pour chaque API que vous souhaitez utiliser

### 4. Créer les identifiants OAuth 2.0

1. Allez sur [Identifiants](https://console.cloud.google.com/apis/credentials?hl=fr&project={{votre nom de project}})
2. Cliquez sur "+ CRÉER DES IDENTIFIANTS" → "ID client OAuth"
3. **Type d'application** : Sélectionnez **"Application de bureau"** (pas "Application Web")
   - ⚠️ Important : Ne choisissez pas "Application Web", sinon l'authentification locale ne fonctionnera pas
4. **Nom** : Donnez un nom descriptif (ex: "Google Auth Wizard Client")
5. Cliquez sur "Créer"
6. **Téléchargez le fichier JSON** :
   - Cliquez sur l'icône de téléchargement à côté de votre client ID
   - Sauvegardez le fichier dans votre répertoire de travail
   - Le fichier sera nommé `client_secret_[ID].apps.googleusercontent.com.json`

### 5. Configuration des URIs de redirection (automatique)

Pour une application de bureau, les URIs de redirection sont gérés automatiquement par Google :
- `http://localhost` (avec port dynamique)
- L'outil utilisera `http://localhost:8080/callback` par défaut

⚠️ **Note** : Si vous avez choisi "Application Web" par erreur, vous devrez :
1. Supprimer l'identifiant créé
2. Recréer un identifiant de type "Application de bureau"

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

#### Configuration initiale (une seule fois)

1. **Créer un projet Google Cloud** : Créez ou sélectionnez un projet sur [Google Cloud Console](https://console.cloud.google.com/)
2. **Activer les APIs** : Activez les APIs nécessaires via https://console.cloud.google.com/apis/dashboard?hl=fr&project={{votre nom de project}}
3. **Créer les identifiants OAuth2** : 
   - Allez sur https://console.cloud.google.com/apis/credentials?hl=fr&project={{votre nom de project}}
   - Créez un identifiant OAuth 2.0 Client ID
   - Téléchargez le fichier JSON client secret
   - Configurez l'URI de redirection : `http://localhost:8080/callback`

#### Utilisation de l'outil

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

## 🔧 Dépannage

### Erreur: "redirect_uri_mismatch"

**Cause** : Vous avez créé un client OAuth de type "Application Web" au lieu de "Application de bureau".

**Solution** :
1. Supprimez l'identifiant OAuth actuel
2. Créez un nouveau client OAuth de type **"Application de bureau"**
3. Téléchargez le nouveau fichier JSON

### Le navigateur ne s'ouvre pas automatiquement

**Solution** : Copiez l'URL affichée dans le terminal et collez-la manuellement dans votre navigateur.

### Erreur: "Port already in use"

**Cause** : Le port 8080 est déjà utilisé par une autre application.

**Solution** : L'outil essaiera automatiquement jusqu'à 10 ports différents. Si le problème persiste, modifiez `defaultPort` dans `config.yaml`.

### Erreur: "Access blocked: This app's request is invalid"

**Cause** : L'écran de consentement OAuth n'est pas correctement configuré ou votre compte n'est pas ajouté comme utilisateur test.

**Solution** :
1. Vérifiez que l'écran de consentement OAuth est configuré
2. Ajoutez votre adresse e-mail dans les "Utilisateurs test" si l'app n'est pas publiée
3. Assurez-vous que les APIs sont activées dans votre projet

### Erreur: "Token expired"

**Cause** : Les tokens d'accès Google expirent généralement après 1 heure.

**Solution** : Relancez l'outil pour obtenir un nouveau token.

### Impossible de récupérer les scopes

**Cause** : Problème de connexion à Google OAuth Playground ou timeout.

**Solution** :
1. Vérifiez votre connexion Internet
2. Augmentez `scopeTimeout` dans `config.yaml`
3. Réessayez plus tard

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
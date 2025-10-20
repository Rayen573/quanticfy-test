# Quanticfy Test Technique - Analyse Client et Chiffre d'Affaires

Ce projet implémente un algorithme d'analyse de données e-commerce, structuré selon l'approche **Load - Compute in Memory - Export**, pour identifier les meilleurs clients (ceux qui ont généré le plus de chiffre d'affaires) au sein du premier quantile de revenu.

## 🎯 Objectif du Projet

L'objectif principal est de :
1.  **LOAD** : Charger en mémoire les données clients, événements d'achat (depuis le 01/04/2020) et prix des contenus depuis une base MySQL.
2.  **TREAT** : Calculer le chiffre d'affaires (CA) total par client et déterminer les **Top Clients** (ceux du premier quantile de revenu, par défaut les 2.5% les plus élevés). Calculer et afficher des statistiques sur la répartition du CA par quantile.
3.  **EXPORT** : Sauvegarder les Top Clients (`CustomerID`, `Email`, `CA`) dans une table de base de données journalière (`test_export_YYYYMMDD`).

## 🛠️ Technologies Utilisées

* **Langage** : Go (Golang)
* **Base de Données** : MySQL
* **Dépendances notables** :
    * `github.com/go-sql-driver/mysql` : Driver MySQL.
    * `github.com/schollz/progressbar/v3` : Pour afficher la barre de progression (selon la consigne).
    * `github.com/joho/godotenv` : Pour charger les variables d'environnement.

## ⚙️ Configuration et Exécution

### 1. Configuration de l'environnement

Le projet utilise les **variables d'environnement** pour la configuration de la base de données, conformément aux consignes.

Créez un fichier `.env` à la racine du projet ou configurez ces variables dans votre environnement système 


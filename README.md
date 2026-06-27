# Minimalist Paint War
A "Regular Show: Paint War" based minimalist game implemented for Distributed Systems class.

# Design Prototype
<img width="1014" height="773" alt="image" src="https://github.com/user-attachments/assets/e9aff586-def0-4538-8026-102d77770e5d" />

# Documentation
## INTRODUCTION
This document presents the technical proposal for the development of the "Paint War" project.
The system will consist of a distributed entertainment platform, operating in real time, based on an authoritative client-server architecture. The main objective is to ensure state synchronization and data integrity across multiple network-connected agents.

## PLANNED MECHANICS
The project will implement the following rules and features:
Movement: The user will control an avatar with 8-directional movement. Movement logic will be processed on the server to mitigate positioning discrepancies.
Aiming and Shooting: The system will use a fixed aiming reticle in front of the player. Shooting will result in a projectile with a straight-line trajectory.
Objective (Flag Capture): The environment will contain static flags. The collision of a projectile with a flag will change the object's ownership to the corresponding team.
Combat and Vitality System: Each player will have 3 health points (HP). Upon reaching the critical state (0 HP), the server will remove the agent from the simulation for 3 seconds, repositioning it at its team's base.
Victory Condition: Matches will be limited to 120 seconds. The team holding possession of the majority of the flags when the timer ends will be declared the winner.

## SYSTEM ARCHITECTURE
The system will use the Authoritative Server model. Below is the visual representation of the planned infrastructure:

<img width="992" height="561" alt="image" src="https://github.com/user-attachments/assets/42d27c09-f55a-489c-a0c7-e7c33e25a87f" />

## NAVIGATION FLOW AND STATES
The user's progression through the system will follow the logical flow below:

<img width="444" height="476" alt="image" src="https://github.com/user-attachments/assets/850b124a-a4a9-493d-b5e3-e47bf9a520a9" />

## DATA PERSISTENCE
To ensure continuity and the recording of historical information, the project will integrate a persistence layer using PostgreSQL.
Purpose: Store match history, player statistics (wins/losses), and flag capture records.
Integration: The Go backend will use the pgx driver or a lightweight ORM to perform asynchronous write operations at the end of each match, ensuring the database does not block the main game processing loop.
Security: Credentials and connection strings will be managed via environment variables, following Twelve-Factor App best practices.

## INTERFACE AND VISUAL ASPECTS
Rendering will be performed via HTML5 Canvas. The adopted aesthetic will follow high-contrast standards:
Graphic Elements: Players will be represented by colored polygons (Blue/Red). Flags will use pole icons that change their fill according to the owning team.
Heads-Up Display (HUD): Overlay layer via Svelte containing a timer, flag counter, and the local player's vitality status.

## DISTRIBUTED TECHNOLOGIES AND CONCEPTS
Stack: Go (Backend), SvelteJS (Frontend), WebSockets (Communication), PostgreSQL (Persistence).
Synchronization: Use of a fixed 20Hz Tick Rate for transmitting the global world state.
Concurrency: Management of multiple simultaneous connections via Goroutines and state synchronization via Mutexes or Channels to avoid Race Conditions.
Serialization: Use of JSON for message exchange, ensuring portability between frontend and backend technologies.

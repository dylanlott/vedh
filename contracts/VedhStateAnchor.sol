// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract VedhStateAnchor {
    enum CommitType {
        TURN_END,
        STACK_RESOLVE
    }

    struct CommitInput {
        address player;
        bytes32 boardstateHash;
        bytes signature;
    }

    struct GameMeta {
        address[] players;
        uint64 epoch;
        bytes32 lastGameStateHash;
        uint64 lastCommitBlock;
        bool exists;
    }

    bytes32 private constant COMMIT_TYPEHASH =
        keccak256(
            "Commit(bytes32 gameId,uint64 epoch,uint8 commitType,bytes32 gameStateHash,bytes32 boardstateHash,uint64 nonce)"
        );

    bytes32 private immutable DOMAIN_SEPARATOR;
    uint256 private immutable DOMAIN_CHAIN_ID;

    address public owner;

    mapping(bytes32 => GameMeta) public games;
    mapping(bytes32 => mapping(address => uint64)) public nonces;
    mapping(bytes32 => mapping(address => bytes32)) public lastPlayerCommit;

    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event GameRegistered(bytes32 indexed gameId, address[] players);
    event CommitMade(
        bytes32 indexed gameId,
        uint64 indexed epoch,
        CommitType commitType,
        bytes32 gameStateHash,
        address indexed player,
        bytes32 playerCommit,
        uint64 nonce,
        uint256 blockTimestamp
    );

    modifier onlyOwner() {
        require(msg.sender == owner, "not owner");
        _;
    }

    constructor(address initialOwner) {
        require(initialOwner != address(0), "invalid owner");
        owner = initialOwner;
        emit OwnershipTransferred(address(0), initialOwner);

        DOMAIN_CHAIN_ID = block.chainid;
        DOMAIN_SEPARATOR = _buildDomainSeparator();
    }

    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "invalid owner");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }

    function registerGame(bytes32 gameId, address[] calldata players) external onlyOwner {
        require(gameId != bytes32(0), "invalid gameId");
        require(players.length > 0, "no players");
        require(!games[gameId].exists, "game exists");

        games[gameId] = GameMeta({
            players: players,
            epoch: 0,
            lastGameStateHash: bytes32(0),
            lastCommitBlock: 0,
            exists: true
        });

        emit GameRegistered(gameId, players);
    }

    function commitState(
        bytes32 gameId,
        CommitType commitType,
        bytes32 gameStateHash,
        CommitInput[] calldata commits
    ) external onlyOwner {
        GameMeta storage g = games[gameId];
        require(g.exists, "game not found");
        require(commits.length == g.players.length, "invalid commit count");

        uint64 epoch = g.epoch;

        for (uint256 i = 0; i < commits.length; i++) {
            CommitInput calldata c = commits[i];
            require(c.player == g.players[i], "commit order mismatch");

            uint64 nonce = nonces[gameId][c.player];
            bytes32 structHash = keccak256(
                abi.encode(
                    COMMIT_TYPEHASH,
                    gameId,
                    epoch,
                    uint8(commitType),
                    gameStateHash,
                    c.boardstateHash,
                    nonce
                )
            );

            bytes32 digest = _hashTypedDataV4(structHash);
            address recovered = _recover(digest, c.signature);
            require(recovered == c.player, "invalid signature");

            bytes32 playerCommit = keccak256(
                abi.encodePacked(c.boardstateHash, gameStateHash, nonce)
            );

            lastPlayerCommit[gameId][c.player] = playerCommit;
            nonces[gameId][c.player] = nonce + 1;

            emit CommitMade(
                gameId,
                epoch,
                commitType,
                gameStateHash,
                c.player,
                playerCommit,
                nonce,
                block.timestamp
            );
        }

        g.lastGameStateHash = gameStateHash;
        g.epoch = epoch + 1;
        g.lastCommitBlock = uint64(block.number);
    }

    function getPlayers(bytes32 gameId) external view returns (address[] memory) {
        return games[gameId].players;
    }

    function domainSeparator() external view returns (bytes32) {
        if (block.chainid == DOMAIN_CHAIN_ID) {
            return DOMAIN_SEPARATOR;
        }
        return _buildDomainSeparator();
    }

    function commitDigest(
        bytes32 gameId,
        uint64 epoch,
        CommitType commitType,
        bytes32 gameStateHash,
        bytes32 boardstateHash,
        uint64 nonce
    ) external view returns (bytes32) {
        bytes32 structHash = keccak256(
            abi.encode(
                COMMIT_TYPEHASH,
                gameId,
                epoch,
                uint8(commitType),
                gameStateHash,
                boardstateHash,
                nonce
            )
        );
        return _hashTypedDataV4(structHash);
    }

    function _hashTypedDataV4(bytes32 structHash) internal view returns (bytes32) {
        return keccak256(abi.encodePacked("\x19\x01", _domainSeparatorV4(), structHash));
    }

    function _domainSeparatorV4() internal view returns (bytes32) {
        if (block.chainid == DOMAIN_CHAIN_ID) {
            return DOMAIN_SEPARATOR;
        }
        return _buildDomainSeparator();
    }

    function _buildDomainSeparator() internal view returns (bytes32) {
        return
            keccak256(
                abi.encode(
                    keccak256(
                        "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
                    ),
                    keccak256(bytes("VedhStateAnchor")),
                    keccak256(bytes("1")),
                    block.chainid,
                    address(this)
                )
            );
    }

    function _recover(bytes32 digest, bytes calldata signature) internal pure returns (address) {
        if (signature.length != 65) {
            return address(0);
        }

        bytes32 r;
        bytes32 s;
        uint8 v;
        // solhint-disable-next-line no-inline-assembly
        assembly {
            r := calldataload(signature.offset)
            s := calldataload(add(signature.offset, 32))
            v := byte(0, calldataload(add(signature.offset, 64)))
        }

        if (v < 27) {
            v += 27;
        }
        if (v != 27 && v != 28) {
            return address(0);
        }

        // Enforce low-s malleability per EIP-2.
        if (
            uint256(s) >
            0x7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0
        ) {
            return address(0);
        }

        return ecrecover(digest, v, r, s);
    }
}

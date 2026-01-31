// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {VedhStateAnchor} from "../contracts/VedhStateAnchor.sol";

interface Vm {
    function addr(uint256 privateKey) external returns (address);
    function sign(uint256 privateKey, bytes32 digest)
        external
        returns (uint8 v, bytes32 r, bytes32 s);
}

contract VedhStateAnchorTest {
    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    function testCommitStateHappyPath() external {
        VedhStateAnchor anchor = new VedhStateAnchor(address(this));

        uint256 pk1 = 0xA11CE;
        uint256 pk2 = 0xB0B;
        address p1 = vm.addr(pk1);
        address p2 = vm.addr(pk2);

        bytes32 gameId = keccak256("game-1");
        address[] memory players = new address[](2);
        players[0] = p1;
        players[1] = p2;
        anchor.registerGame(gameId, players);

        bytes32 gameStateHash = keccak256("game-state");
        bytes32 b1 = keccak256("board-1");
        bytes32 b2 = keccak256("board-2");

        VedhStateAnchor.CommitInput[] memory commits = new VedhStateAnchor.CommitInput[](2);

        bytes32 d1 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            b1,
            0
        );
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(pk1, d1);
        commits[0] = VedhStateAnchor.CommitInput({
            player: p1,
            boardstateHash: b1,
            signature: abi.encodePacked(r1, s1, v1)
        });

        bytes32 d2 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            b2,
            0
        );
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(pk2, d2);
        commits[1] = VedhStateAnchor.CommitInput({
            player: p2,
            boardstateHash: b2,
            signature: abi.encodePacked(r2, s2, v2)
        });

        anchor.commitState(gameId, VedhStateAnchor.CommitType.TURN_END, gameStateHash, commits);

        require(anchor.nonces(gameId, p1) == 1, "nonce p1 not incremented");
        require(anchor.nonces(gameId, p2) == 1, "nonce p2 not incremented");

        bytes32 expected1 = keccak256(abi.encodePacked(b1, gameStateHash, uint64(0)));
        bytes32 expected2 = keccak256(abi.encodePacked(b2, gameStateHash, uint64(0)));
        require(anchor.lastPlayerCommit(gameId, p1) == expected1, "commit p1 mismatch");
        require(anchor.lastPlayerCommit(gameId, p2) == expected2, "commit p2 mismatch");
    }

    function testCommitOrderMismatchReverts() external {
        VedhStateAnchor anchor = new VedhStateAnchor(address(this));

        uint256 pk1 = 0xA11CE;
        uint256 pk2 = 0xB0B;
        address p1 = vm.addr(pk1);
        address p2 = vm.addr(pk2);

        bytes32 gameId = keccak256("game-2");
        address[] memory players = new address[](2);
        players[0] = p1;
        players[1] = p2;
        anchor.registerGame(gameId, players);

        bytes32 gameStateHash = keccak256("game-state");
        bytes32 b1 = keccak256("board-1");
        bytes32 b2 = keccak256("board-2");

        VedhStateAnchor.CommitInput[] memory commits = new VedhStateAnchor.CommitInput[](2);

        bytes32 d1 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.STACK_RESOLVE,
            gameStateHash,
            b1,
            0
        );
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(pk1, d1);
        commits[0] = VedhStateAnchor.CommitInput({
            player: p2,
            boardstateHash: b2,
            signature: abi.encodePacked(r1, s1, v1)
        });

        bytes32 d2 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.STACK_RESOLVE,
            gameStateHash,
            b2,
            0
        );
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(pk2, d2);
        commits[1] = VedhStateAnchor.CommitInput({
            player: p1,
            boardstateHash: b1,
            signature: abi.encodePacked(r2, s2, v2)
        });

        bytes memory data = abi.encodeWithSelector(
            anchor.commitState.selector,
            gameId,
            VedhStateAnchor.CommitType.STACK_RESOLVE,
            gameStateHash,
            commits
        );
        (bool ok, ) = address(anchor).call(data);
        require(!ok, "expected revert");
    }

    function testInvalidSignatureReverts() external {
        VedhStateAnchor anchor = new VedhStateAnchor(address(this));

        uint256 pk1 = 0xA11CE;
        uint256 pk2 = 0xB0B;
        address p1 = vm.addr(pk1);
        address p2 = vm.addr(pk2);

        bytes32 gameId = keccak256("game-3");
        address[] memory players = new address[](2);
        players[0] = p1;
        players[1] = p2;
        anchor.registerGame(gameId, players);

        bytes32 gameStateHash = keccak256("game-state");
        bytes32 b1 = keccak256("board-1");
        bytes32 b2 = keccak256("board-2");

        VedhStateAnchor.CommitInput[] memory commits = new VedhStateAnchor.CommitInput[](2);

        bytes32 d1 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            b1,
            0
        );
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(pk2, d1);
        commits[0] = VedhStateAnchor.CommitInput({
            player: p1,
            boardstateHash: b1,
            signature: abi.encodePacked(r1, s1, v1)
        });

        bytes32 d2 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            b2,
            0
        );
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(pk2, d2);
        commits[1] = VedhStateAnchor.CommitInput({
            player: p2,
            boardstateHash: b2,
            signature: abi.encodePacked(r2, s2, v2)
        });

        bytes memory data = abi.encodeWithSelector(
            anchor.commitState.selector,
            gameId,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            commits
        );
        (bool ok, ) = address(anchor).call(data);
        require(!ok, "expected revert");
    }

    function testNonceReplayReverts() external {
        VedhStateAnchor anchor = new VedhStateAnchor(address(this));

        uint256 pk1 = 0xA11CE;
        uint256 pk2 = 0xB0B;
        address p1 = vm.addr(pk1);
        address p2 = vm.addr(pk2);

        bytes32 gameId = keccak256("game-4");
        address[] memory players = new address[](2);
        players[0] = p1;
        players[1] = p2;
        anchor.registerGame(gameId, players);

        bytes32 gameStateHash = keccak256("game-state");
        bytes32 b1 = keccak256("board-1");
        bytes32 b2 = keccak256("board-2");

        VedhStateAnchor.CommitInput[] memory commits = new VedhStateAnchor.CommitInput[](2);

        bytes32 d1 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            b1,
            0
        );
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(pk1, d1);
        commits[0] = VedhStateAnchor.CommitInput({
            player: p1,
            boardstateHash: b1,
            signature: abi.encodePacked(r1, s1, v1)
        });

        bytes32 d2 = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            b2,
            0
        );
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(pk2, d2);
        commits[1] = VedhStateAnchor.CommitInput({
            player: p2,
            boardstateHash: b2,
            signature: abi.encodePacked(r2, s2, v2)
        });

        anchor.commitState(gameId, VedhStateAnchor.CommitType.TURN_END, gameStateHash, commits);

        bytes memory data = abi.encodeWithSelector(
            anchor.commitState.selector,
            gameId,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash,
            commits
        );
        (bool ok, ) = address(anchor).call(data);
        require(!ok, "expected replay revert");
    }

    function testEpochIncrements() external {
        VedhStateAnchor anchor = new VedhStateAnchor(address(this));

        uint256 pk1 = 0xA11CE;
        uint256 pk2 = 0xB0B;
        address p1 = vm.addr(pk1);
        address p2 = vm.addr(pk2);

        bytes32 gameId = keccak256("game-5");
        address[] memory players = new address[](2);
        players[0] = p1;
        players[1] = p2;
        anchor.registerGame(gameId, players);

        bytes32 gameStateHash1 = keccak256("game-state-1");
        bytes32 b1 = keccak256("board-1");
        bytes32 b2 = keccak256("board-2");

        VedhStateAnchor.CommitInput[] memory commits1 = new VedhStateAnchor.CommitInput[](2);

        bytes32 d1a = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash1,
            b1,
            0
        );
        (uint8 v1a, bytes32 r1a, bytes32 s1a) = vm.sign(pk1, d1a);
        commits1[0] = VedhStateAnchor.CommitInput({
            player: p1,
            boardstateHash: b1,
            signature: abi.encodePacked(r1a, s1a, v1a)
        });

        bytes32 d2a = anchor.commitDigest(
            gameId,
            0,
            VedhStateAnchor.CommitType.TURN_END,
            gameStateHash1,
            b2,
            0
        );
        (uint8 v2a, bytes32 r2a, bytes32 s2a) = vm.sign(pk2, d2a);
        commits1[1] = VedhStateAnchor.CommitInput({
            player: p2,
            boardstateHash: b2,
            signature: abi.encodePacked(r2a, s2a, v2a)
        });

        anchor.commitState(gameId, VedhStateAnchor.CommitType.TURN_END, gameStateHash1, commits1);
        (uint64 epochAfterFirst, , , ) = anchor.games(gameId);
        require(epochAfterFirst == 1, "epoch not incremented");

        bytes32 gameStateHash2 = keccak256("game-state-2");
        bytes32 d1b = anchor.commitDigest(
            gameId,
            1,
            VedhStateAnchor.CommitType.STACK_RESOLVE,
            gameStateHash2,
            b1,
            1
        );
        (uint8 v1b, bytes32 r1b, bytes32 s1b) = vm.sign(pk1, d1b);
        VedhStateAnchor.CommitInput[] memory commits2 = new VedhStateAnchor.CommitInput[](2);
        commits2[0] = VedhStateAnchor.CommitInput({
            player: p1,
            boardstateHash: b1,
            signature: abi.encodePacked(r1b, s1b, v1b)
        });

        bytes32 d2b = anchor.commitDigest(
            gameId,
            1,
            VedhStateAnchor.CommitType.STACK_RESOLVE,
            gameStateHash2,
            b2,
            1
        );
        (uint8 v2b, bytes32 r2b, bytes32 s2b) = vm.sign(pk2, d2b);
        commits2[1] = VedhStateAnchor.CommitInput({
            player: p2,
            boardstateHash: b2,
            signature: abi.encodePacked(r2b, s2b, v2b)
        });

        anchor.commitState(gameId, VedhStateAnchor.CommitType.STACK_RESOLVE, gameStateHash2, commits2);
        (uint64 epochAfterSecond, , , ) = anchor.games(gameId);
        require(epochAfterSecond == 2, "epoch not incremented again");
    }
}

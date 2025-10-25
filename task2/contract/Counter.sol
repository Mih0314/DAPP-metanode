pragma solidity ^0.8;

contract Counter {
    uint256 private count;
    constructor(uint256 _count) {
        count = _count;
    }

    function incr() public {
        count++;
    }

    function add(uint256 _num) public {
        count +=_num;
    }

    function getCount()public view returns(uint256) {
        return count;
    }
}
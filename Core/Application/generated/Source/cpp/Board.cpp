/*
    Board.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "Board.h"

std::string Board::getId() {    
    return this->id;
}

void Board::setId(std::string &value) {
    this->id = value;
}

std::string Board::getTitle() {    
    return this->title;
}

void Board::setTitle(std::string &value) {
    this->title = value;
}

std::string Board::getDescription() {    
    return this->description;
}

void Board::setDescription(std::string &value) {
    this->description = value;
}

int Board::getCreated() {    
    return this->created;
}

void Board::setCreated(int &value) {
    this->created = value;
}

int Board::getModified() {    
    return this->modified;
}

void Board::setModified(int &value) {
    this->modified = value;
}

bool Board::isDeleted() {    
    return this->deleted;
}

void Board::setDeleted(bool &value) {
    this->deleted = value;
}


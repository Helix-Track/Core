/*
    TicketBoardMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketBoardMapping.h"

std::string TicketBoardMapping::getId() {    
    return this->id;
}

void TicketBoardMapping::setId(std::string &value) {
    this->id = value;
}

std::string TicketBoardMapping::getTicketId() {    
    return this->ticketId;
}

void TicketBoardMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

std::string TicketBoardMapping::getBoardId() {    
    return this->boardId;
}

void TicketBoardMapping::setBoardId(std::string &value) {
    this->boardId = value;
}

int TicketBoardMapping::getCreated() {    
    return this->created;
}

void TicketBoardMapping::setCreated(int &value) {
    this->created = value;
}

int TicketBoardMapping::getModified() {    
    return this->modified;
}

void TicketBoardMapping::setModified(int &value) {
    this->modified = value;
}

bool TicketBoardMapping::isDeleted() {    
    return this->deleted;
}

void TicketBoardMapping::setDeleted(bool &value) {
    this->deleted = value;
}


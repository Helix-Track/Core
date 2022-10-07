/*
    TicketBoardMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "TicketBoardMappings.h"

std::string TicketBoardMappings::getId() {    
    return this->id;
}

void TicketBoardMappings::setId(std::string value) {
    this->id = value;
}

std::string TicketBoardMappings::getTicketId() {    
    return this->ticketId;
}

void TicketBoardMappings::setTicketId(std::string value) {
    this->ticketId = value;
}

std::string TicketBoardMappings::getBoardId() {    
    return this->boardId;
}

void TicketBoardMappings::setBoardId(std::string value) {
    this->boardId = value;
}

int TicketBoardMappings::getCreated() {    
    return this->created;
}

void TicketBoardMappings::setCreated(int value) {
    this->created = value;
}

int TicketBoardMappings::getModified() {    
    return this->modified;
}

void TicketBoardMappings::setModified(int value) {
    this->modified = value;
}

bool TicketBoardMappings::isDeleted() {    
    return this->deleted;
}

void TicketBoardMappings::setDeleted(bool value) {
    this->deleted = value;
}


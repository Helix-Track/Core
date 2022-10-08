/*
    Ticket.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Ticket.h"

std::string Ticket::getId() {    
    return this->id;
}

void Ticket::setId(std::string value) {
    this->id = value;
}

int Ticket::getTicketNumber() {    
    return this->ticketNumber;
}

void Ticket::setTicketNumber(int value) {
    this->ticketNumber = value;
}

int Ticket::getPosition() {    
    return this->position;
}

void Ticket::setPosition(int value) {
    this->position = value;
}

std::string Ticket::getTitle() {    
    return this->title;
}

void Ticket::setTitle(std::string value) {
    this->title = value;
}

std::string Ticket::getDescription() {    
    return this->description;
}

void Ticket::setDescription(std::string value) {
    this->description = value;
}

int Ticket::getCreated() {    
    return this->created;
}

void Ticket::setCreated(int value) {
    this->created = value;
}

int Ticket::getModified() {    
    return this->modified;
}

void Ticket::setModified(int value) {
    this->modified = value;
}

std::string Ticket::getTicketTypeId() {    
    return this->ticketTypeId;
}

void Ticket::setTicketTypeId(std::string value) {
    this->ticketTypeId = value;
}

std::string Ticket::getTicketStatusId() {    
    return this->ticketStatusId;
}

void Ticket::setTicketStatusId(std::string value) {
    this->ticketStatusId = value;
}

std::string Ticket::getProjectId() {    
    return this->projectId;
}

void Ticket::setProjectId(std::string value) {
    this->projectId = value;
}

std::string Ticket::getUserId() {    
    return this->userId;
}

void Ticket::setUserId(std::string value) {
    this->userId = value;
}

double Ticket::getEstimation() {    
    return this->estimation;
}

void Ticket::setEstimation(double value) {
    this->estimation = value;
}

int Ticket::getStoryPoints() {    
    return this->storyPoints;
}

void Ticket::setStoryPoints(int value) {
    this->storyPoints = value;
}

std::string Ticket::getCreator() {    
    return this->creator;
}

void Ticket::setCreator(std::string value) {
    this->creator = value;
}

bool Ticket::isDeleted() {    
    return this->deleted;
}

void Ticket::setDeleted(bool value) {
    this->deleted = value;
}


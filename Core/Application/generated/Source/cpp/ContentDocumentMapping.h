/*
    ContentDocumentMapping.h
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class ContentDocumentMapping {

private:
    std::string id;
    std::string documentId;
    std::string content;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getDocumentId();
    void setDocumentId(std::string &value);
    std::string getContent();
    void setContent(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};
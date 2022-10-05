/*
    Extensions.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Extensions {

private:
    std::string id;
    int created;
    int modified;
    std::string title;
    std::string description;
    std::string extensionKey;
    bool enabled;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    std::string getTitle();
    void setTitle(std::string value);
    std::string getDescription();
    void setDescription(std::string value);
    std::string getExtensionKey();
    void setExtensionKey(std::string value);
    bool isEnabled();
    void setEnabled(bool value);
    bool isDeleted();
    void setDeleted(bool value);
};
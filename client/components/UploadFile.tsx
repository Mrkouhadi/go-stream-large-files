"use client";

// export default UploadFile;
import React, { useState } from "react";

const UploadFile = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadProgress, setUploadProgress] = useState<number>(0);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    const formData = new FormData();
    formData.append("file", selectedFile);

    const xhr = new XMLHttpRequest();

    // Track upload progress
    xhr.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable) {
        const progress = Math.round((event.loaded / event.total) * 100);
        setUploadProgress(progress);
      }
    });

    xhr.onreadystatechange = () => {
      if (xhr.readyState === XMLHttpRequest.DONE) {
        if (xhr.status === 200) {
          console.log("File uploaded successfully!");
          setUploadProgress(0);
        } else {
          console.error("Error uploading file:", xhr.statusText);
          alert("Failed to upload file. Please try again.");
        }
      }
    };

    xhr.open("POST", "http://localhost:8080/upload", true);
    xhr.send(formData);
  };
  return (
    <div>
      <input type="file" onChange={handleFileChange} />
      <button onClick={handleUpload}>Upload</button>
      {uploadProgress > 0 && (
        <progress
          className="bg-blue-400 ml-8 h-6 w-64  "
          value={uploadProgress}
          max="100"
        />
      )}
    </div>
  );
};

export default UploadFile;

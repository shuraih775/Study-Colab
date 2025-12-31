package com.studycollab.file_service.service;

import com.studycollab.file_service.config.MinioConfig;
import com.studycollab.file_service.proto.*;
import io.minio.GetPresignedObjectUrlArgs;
import io.minio.MinioClient;
import io.minio.http.Method;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
public class MinioFileService {

    private final MinioClient publicClient;
    private final MinioConfig config;

    public MinioFileService(
            @Qualifier("publicMinio") MinioClient publicClient,
            MinioConfig config
    ) {
        this.publicClient = publicClient;
        this.config = config;
    }

    public UploadResponse generateUpload(UploadRequest req) {
        String fileId = UUID.randomUUID().toString();

        try {
            String url = publicClient.getPresignedObjectUrl(
                    GetPresignedObjectUrlArgs.builder()
                            .method(Method.PUT)
                            .bucket(config.getBucket())
                            .object(fileId)
                            .build()
            );

            return UploadResponse.newBuilder()
                    .setFileId(fileId)
                    .setPresignedUrl(url)
                    .build();

        } catch (Exception e) {
            throw new RuntimeException("Failed to generate upload URL", e);
        }
    }

    public DownloadResponse generateDownload(DownloadRequest req) {
        try {
            String url = publicClient.getPresignedObjectUrl(
                    GetPresignedObjectUrlArgs.builder()
                            .method(Method.GET)
                            .bucket(config.getBucket())
                            .object(req.getFileId())
                            .build()
            );

            return DownloadResponse.newBuilder()
                    .setPresignedUrl(url)
                    .build();

        } catch (Exception e) {
            throw new RuntimeException("Failed to generate download URL", e);
        }
    }
}
